package bedrock

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"nu/internal/provider"
)

const (
	defaultRegion            = "us-east-1"
	maxEventStreamFrameBytes = 8 * 1024 * 1024
)

// Credentials are AWS credentials used for Bedrock runtime signing.
type Credentials struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	Region          string
}

// Config configures a Bedrock ConverseStream client.
type Config struct {
	BaseURL     string
	Credentials Credentials
	Client      *http.Client
	Now         func() time.Time
}

// Client implements Bedrock ConverseStream.
type Client struct {
	baseURL     string
	credentials Credentials
	client      *http.Client
	now         func() time.Time
}

// New constructs a Bedrock client.
func New(cfg Config) *Client {
	if cfg.Credentials.Region == "" {
		cfg.Credentials.Region = defaultRegion
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://bedrock-runtime." + cfg.Credentials.Region + ".amazonaws.com"
	}
	if cfg.Client == nil {
		cfg.Client = http.DefaultClient
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	return &Client{
		baseURL:     strings.TrimRight(cfg.BaseURL, "/"),
		credentials: cfg.Credentials,
		client:      cfg.Client,
		now:         cfg.Now,
	}
}

// BuildConversePayload builds a Bedrock ConverseStream request body.
func BuildConversePayload(req provider.Request) (map[string]any, error) {
	if err := provider.ValidateRequest(req); err != nil {
		return nil, err
	}
	messages := make([]map[string]any, 0, len(req.Messages))
	for _, message := range req.Messages {
		messages = append(messages, converseMessage(message))
	}
	return map[string]any{"messages": messages}, nil
}

func converseMessage(message provider.Message) map[string]any {
	if message.Role == "assistant" && message.ToolCallID != "" {
		return map[string]any{
			"role": "assistant",
			"content": []map[string]any{{
				"toolUse": map[string]any{
					"toolUseId": message.ToolCallID,
					"name":      message.Name,
					"input":     decodeJSONOrText(message.Content),
				},
			}},
		}
	}
	if message.Role == "tool" {
		return map[string]any{
			"role": "user",
			"content": []map[string]any{{
				"toolResult": map[string]any{
					"toolUseId": message.ToolCallID,
					"content":   []map[string]string{{"text": message.Content}},
				},
			}},
		}
	}
	return map[string]any{
		"role":    message.Role,
		"content": []map[string]string{{"text": message.Content}},
	}
}

// Sign applies AWS Signature Version 4 headers to req.
func Sign(req *http.Request, body []byte, creds Credentials, now time.Time) error {
	if creds.AccessKeyID == "" || creds.SecretAccessKey == "" {
		return fmt.Errorf("bedrock: missing aws credentials")
	}
	if creds.Region == "" {
		creds.Region = defaultRegion
	}
	amzDate := now.UTC().Format("20060102T150405Z")
	dateStamp := now.UTC().Format("20060102")
	payloadHash := sha256Hex(body)
	req.Header.Set("X-Amz-Date", amzDate)
	req.Header.Set("X-Amz-Content-Sha256", payloadHash)
	if creds.SessionToken != "" {
		req.Header.Set("X-Amz-Security-Token", creds.SessionToken)
	}

	canonicalHeaders, signedHeaders := canonicalHeaders(req)
	scope := dateStamp + "/" + creds.Region + "/bedrock/aws4_request"
	canonicalRequest := strings.Join([]string{
		req.Method,
		emptyDefault(req.URL.EscapedPath(), "/"),
		req.URL.RawQuery,
		canonicalHeaders,
		signedHeaders,
		payloadHash,
	}, "\n")
	stringToSign := strings.Join([]string{
		"AWS4-HMAC-SHA256",
		amzDate,
		scope,
		sha256Hex([]byte(canonicalRequest)),
	}, "\n")
	signature := hex.EncodeToString(hmacSHA256(signingKey(creds.SecretAccessKey, dateStamp, creds.Region), []byte(stringToSign)))
	req.Header.Set(
		"Authorization",
		"AWS4-HMAC-SHA256 Credential="+creds.AccessKeyID+"/"+scope+", SignedHeaders="+signedHeaders+", Signature="+signature,
	)
	return nil
}

// Stream starts one Bedrock ConverseStream request.
func (c *Client) Stream(ctx context.Context, req provider.Request) (<-chan provider.Event, error) {
	payload, err := BuildConversePayload(req)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal bedrock request: %w", err)
	}
	endpoint := c.baseURL + "/model/" + escapeModelID(req.Model) + "/converse-stream"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("create bedrock request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/vnd.amazon.eventstream")
	if err := Sign(httpReq, data, c.credentials, c.now()); err != nil {
		return nil, err
	}
	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send bedrock request: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("bedrock http %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	ch := make(chan provider.Event)
	go func() {
		defer close(ch)
		defer resp.Body.Close()
		send(ctx, ch, provider.Event{Type: provider.EventStart, Provider: req.Provider, API: req.API, Model: req.Model})
		if err := parseEventStream(ctx, resp.Body, ch); err != nil {
			send(ctx, ch, provider.Event{Type: provider.EventError, ErrorClass: "fatal", Message: err.Error()})
		}
	}()
	return ch, nil
}

func parseEventStream(ctx context.Context, r io.Reader, ch chan<- provider.Event) error {
	started := map[int]bool{}
	for {
		payload, err := readEventStreamPayload(r)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("read bedrock stream: %w", err)
		}
		if err := handlePayload(ctx, ch, started, payload); err != nil {
			return err
		}
	}
}

func handlePayload(ctx context.Context, ch chan<- provider.Event, started map[int]bool, payload []byte) error {
	var ev map[string]json.RawMessage
	if err := json.Unmarshal(payload, &ev); err != nil {
		return fmt.Errorf("decode bedrock payload: %w", err)
	}
	if raw := ev["contentBlockStart"]; raw != nil {
		var start struct {
			ContentBlockIndex int `json:"contentBlockIndex"`
			Start             struct {
				ToolUse struct {
					ToolUseID string `json:"toolUseId"`
					Name      string `json:"name"`
				} `json:"toolUse"`
			} `json:"start"`
		}
		if err := json.Unmarshal(raw, &start); err != nil {
			return fmt.Errorf("decode bedrock block start: %w", err)
		}
		if start.Start.ToolUse.ToolUseID != "" {
			started[start.ContentBlockIndex] = true
			send(ctx, ch, provider.Event{
				Type:       provider.EventToolCallStart,
				Index:      start.ContentBlockIndex,
				ToolCallID: start.Start.ToolUse.ToolUseID,
				ToolName:   start.Start.ToolUse.Name,
			})
		}
	}
	if raw := ev["contentBlockDelta"]; raw != nil {
		var delta struct {
			ContentBlockIndex int `json:"contentBlockIndex"`
			Delta             struct {
				Text    string `json:"text"`
				ToolUse struct {
					Input string `json:"input"`
				} `json:"toolUse"`
			} `json:"delta"`
		}
		if err := json.Unmarshal(raw, &delta); err != nil {
			return fmt.Errorf("decode bedrock block delta: %w", err)
		}
		if delta.Delta.Text != "" {
			send(ctx, ch, provider.Event{Type: provider.EventText, Index: delta.ContentBlockIndex, Delta: delta.Delta.Text})
		}
		if delta.Delta.ToolUse.Input != "" {
			send(ctx, ch, provider.Event{
				Type:  provider.EventToolCallDelta,
				Index: delta.ContentBlockIndex,
				Delta: delta.Delta.ToolUse.Input,
			})
		}
	}
	if raw := ev["contentBlockStop"]; raw != nil {
		var stop struct {
			ContentBlockIndex int `json:"contentBlockIndex"`
		}
		if err := json.Unmarshal(raw, &stop); err != nil {
			return fmt.Errorf("decode bedrock block stop: %w", err)
		}
		if started[stop.ContentBlockIndex] {
			send(ctx, ch, provider.Event{Type: provider.EventToolCallEnd, Index: stop.ContentBlockIndex})
		}
	}
	if raw := ev["messageStop"]; raw != nil {
		var stop struct {
			StopReason string `json:"stopReason"`
		}
		if err := json.Unmarshal(raw, &stop); err != nil {
			return fmt.Errorf("decode bedrock message stop: %w", err)
		}
		send(ctx, ch, provider.Event{Type: provider.EventDone, StopReason: normalizeBedrockStop(stop.StopReason)})
	}
	return nil
}

func readEventStreamPayload(r io.Reader) ([]byte, error) {
	var prelude [12]byte
	if _, err := io.ReadFull(r, prelude[:]); err != nil {
		return nil, err
	}
	totalLen := binary.BigEndian.Uint32(prelude[0:4])
	headersLen := binary.BigEndian.Uint32(prelude[4:8])
	if crc32.ChecksumIEEE(prelude[0:8]) != binary.BigEndian.Uint32(prelude[8:12]) {
		return nil, fmt.Errorf("bedrock event prelude crc mismatch")
	}
	if totalLen < 16 || headersLen > totalLen-16 {
		return nil, fmt.Errorf("bedrock event length is invalid")
	}
	// The remote frame length is capped before allocation to bound memory use.
	if totalLen > maxEventStreamFrameBytes {
		return nil, fmt.Errorf("bedrock event frame too large: %d bytes", totalLen)
	}
	rest := make([]byte, totalLen-12)
	if _, err := io.ReadFull(r, rest); err != nil {
		return nil, err
	}
	message := append(prelude[:], rest...)
	wantCRC := binary.BigEndian.Uint32(message[len(message)-4:])
	if crc32.ChecksumIEEE(message[:len(message)-4]) != wantCRC {
		return nil, fmt.Errorf("bedrock event message crc mismatch")
	}
	payloadStart := int(12 + headersLen)
	payloadEnd := len(message) - 4
	return message[payloadStart:payloadEnd], nil
}

func decodeJSONOrText(raw string) any {
	var value any
	if err := json.Unmarshal([]byte(raw), &value); err == nil {
		return value
	}
	return map[string]string{"text": raw}
}

func canonicalHeaders(req *http.Request) (string, string) {
	values := map[string]string{"host": req.URL.Host}
	for name, headerValues := range req.Header {
		lower := strings.ToLower(name)
		values[lower] = strings.Join(headerValues, ",")
	}
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	sort.Strings(names)
	var lines strings.Builder
	for _, name := range names {
		lines.WriteString(name)
		lines.WriteString(":")
		lines.WriteString(strings.TrimSpace(values[name]))
		lines.WriteString("\n")
	}
	return lines.String(), strings.Join(names, ";")
}

func signingKey(secret, dateStamp, region string) []byte {
	dateKey := hmacSHA256([]byte("AWS4"+secret), []byte(dateStamp))
	regionKey := hmacSHA256(dateKey, []byte(region))
	serviceKey := hmacSHA256(regionKey, []byte("bedrock"))
	return hmacSHA256(serviceKey, []byte("aws4_request"))
}

func hmacSHA256(key, data []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return mac.Sum(nil)
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func emptyDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func escapeModelID(modelID string) string {
	escaped := url.PathEscape(modelID)
	// Bedrock examples keep ':' in the model id path segment; keep that readable while
	// still escaping '/' for ARNs and inference profile identifiers.
	return strings.ReplaceAll(escaped, "%3A", ":")
}

func normalizeBedrockStop(reason string) string {
	switch reason {
	case "tool_use":
		return "tool_use"
	case "max_tokens":
		return "length"
	case "guardrail_intervened":
		return "content_filter"
	default:
		return "stop"
	}
}

func send(ctx context.Context, ch chan<- provider.Event, ev provider.Event) bool {
	select {
	case <-ctx.Done():
		return false
	case ch <- ev:
		return true
	}
}
