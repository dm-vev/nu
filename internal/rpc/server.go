package rpc

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"sync"

	"nu/internal/agent"
	"nu/internal/tool/bash"
)

const defaultBashOutputLimit = 16 * 1024

var errStop = errors.New("rpc stop")

// Options configures one RPC server.
type Options struct {
	Stdin      io.Reader
	Stdout     io.Writer
	Stderr     io.Writer
	CWD        string
	SessionID  string
	Provider   string
	API        string
	Model      string
	ModelLabel string
}

// Server owns one JSONL RPC session.
type Server struct {
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer

	writeMu  sync.Mutex
	writeErr error

	mu               sync.Mutex
	agent            *agent.Agent
	cwd              string
	sessionID        string
	sessionName      string
	provider         string
	api              string
	model            string
	modelLabel       string
	thinkingLevel    string
	steeringMode     string
	followUpMode     string
	autoCompaction   bool
	autoRetry        bool
	running          bool
	shutdown         bool
	settings         map[string]any
	steeringQueue    []string
	followUpQueue    []string
	messages         []rpcMessage
	nextMessageIndex int

	wg sync.WaitGroup
}

type rpcMessage struct {
	ID      string `json:"id"`
	Role    string `json:"role"`
	Content string `json:"content"`
}

type commandEnvelope struct {
	ID                     string         `json:"id,omitempty"`
	Type                   string         `json:"type"`
	Message                string         `json:"message,omitempty"`
	StreamingBehavior      string         `json:"streaming_behavior,omitempty"`
	StreamingBehaviorCamel string         `json:"streamingBehavior,omitempty"`
	ParentSession          string         `json:"parent_session,omitempty"`
	ParentSessionCamel     string         `json:"parentSession,omitempty"`
	Provider               string         `json:"provider,omitempty"`
	ModelID                string         `json:"modelId,omitempty"`
	ModelIDSnake           string         `json:"model_id,omitempty"`
	Level                  string         `json:"level,omitempty"`
	Mode                   string         `json:"mode,omitempty"`
	CustomInstructions     string         `json:"customInstructions,omitempty"`
	CustomInstructionsAlt  string         `json:"custom_instructions,omitempty"`
	Enabled                bool           `json:"enabled,omitempty"`
	CommandText            string         `json:"command,omitempty"`
	ExcludeFromContext     bool           `json:"excludeFromContext,omitempty"`
	OutputPath             string         `json:"outputPath,omitempty"`
	SessionPath            string         `json:"sessionPath,omitempty"`
	EntryID                string         `json:"entryId,omitempty"`
	Since                  string         `json:"since,omitempty"`
	Name                   string         `json:"name,omitempty"`
	Settings               map[string]any `json:"settings,omitempty"`
	Persist                bool           `json:"persist,omitempty"`
}

type response map[string]any

// NewServer creates an idle JSONL RPC server.
func NewServer(opts Options) *Server {
	if opts.Stdin == nil {
		opts.Stdin = strings.NewReader("")
	}
	if opts.Stdout == nil {
		opts.Stdout = io.Discard
	}
	if opts.Stderr == nil {
		opts.Stderr = io.Discard
	}
	sessionID := strings.TrimSpace(opts.SessionID)
	if sessionID == "" {
		sessionID = newID("session")
	}
	return &Server{
		stdin:          opts.Stdin,
		stdout:         opts.Stdout,
		stderr:         opts.Stderr,
		cwd:            opts.CWD,
		sessionID:      sessionID,
		provider:       firstNonEmpty(opts.Provider, "test"),
		api:            firstNonEmpty(opts.API, "test"),
		model:          firstNonEmpty(opts.Model, "test"),
		modelLabel:     firstNonEmpty(opts.ModelLabel, opts.Model, "test"),
		thinkingLevel:  "off",
		steeringMode:   "all",
		followUpMode:   "all",
		autoCompaction: true,
		autoRetry:      true,
		settings:       map[string]any{},
	}
}

// SetAgent injects the provider-backed agent after server construction.
func (s *Server) SetAgent(a *agent.Agent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.agent = a
}

// Emit forwards agent events to RPC stdout and updates lightweight state.
func (s *Server) Emit(ev agent.Event) {
	s.recordEvent(ev)
	if err := s.write(ev); err != nil {
		s.setWriteErr(err)
	}
}

// Run serves JSONL commands until EOF, shutdown, context cancellation, or write failure.
func (s *Server) Run(ctx context.Context) error {
	err := ReadLines(s.stdin, func(line string) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := s.handleLine(ctx, line); err != nil {
			return err
		}
		if s.shouldStop() {
			return errStop
		}
		return s.currentWriteErr()
	})
	if errors.Is(err, errStop) {
		err = nil
	}
	if err != nil {
		return err
	}
	if !s.shouldStop() {
		// EOF on stdin is a headless-client shutdown signal, matching Pi's RPC mode.
		s.requestShutdown()
	}
	if err := s.waitIdle(ctx); err != nil {
		return err
	}
	return s.currentWriteErr()
}

func (s *Server) handleLine(ctx context.Context, line string) error {
	var command commandEnvelope
	if err := json.Unmarshal([]byte(line), &command); err != nil {
		return s.write(failure("", "parse", fmt.Sprintf("failed to parse command: %v", err)))
	}
	if command.Type == "extension_ui_response" {
		return nil
	}

	resp, ok := s.handleCommand(ctx, command)
	if !ok {
		return nil
	}
	return s.write(resp)
}

func (s *Server) handleCommand(ctx context.Context, command commandEnvelope) (response, bool) {
	switch command.Type {
	case "prompt":
		behavior := command.StreamingBehavior
		if behavior == "" {
			behavior = command.StreamingBehaviorCamel
		}
		if s.isRunning() {
			switch behavior {
			case "steer":
				s.enqueueSteering(command.Message)
				return success(command.ID, "prompt", nil), true
			case "follow_up", "followUp":
				s.enqueueFollowUp(command.Message)
				return success(command.ID, "prompt", nil), true
			default:
				return failure(command.ID, "prompt", "agent busy; use streaming_behavior steer or follow_up"), true
			}
		}
		return s.startPrompt(ctx, command.ID, command.Message), true
	case "steer":
		s.enqueueSteering(command.Message)
		return success(command.ID, "steer", nil), true
	case "follow_up":
		s.enqueueFollowUp(command.Message)
		return success(command.ID, "follow_up", nil), true
	case "abort":
		s.abort()
		return success(command.ID, "abort", nil), true
	case "new_session":
		parent := firstNonEmpty(command.ParentSession, command.ParentSessionCamel)
		s.newSession(parent)
		return success(command.ID, "new_session", map[string]any{"cancelled": false}), true
	case "get_state", "state":
		return success(command.ID, command.Type, s.stateData()), true
	case "set_settings":
		s.setSettings(command.Settings, command.Persist)
		return success(command.ID, "set_settings", s.settingsData()), true
	case "set_model":
		modelID := firstNonEmpty(command.ModelID, command.ModelIDSnake)
		if modelID == "" {
			return failure(command.ID, "set_model", "missing model id"), true
		}
		if err := s.setModel(command.Provider, modelID); err != nil {
			return failure(command.ID, "set_model", err.Error()), true
		}
		return success(command.ID, "set_model", s.modelData()), true
	case "cycle_model":
		return success(command.ID, "cycle_model", map[string]any{
			"model":         s.modelData(),
			"thinkingLevel": s.thinking(),
			"isScoped":      false,
		}), true
	case "get_available_models":
		return success(command.ID, "get_available_models", map[string]any{"models": []any{s.modelData()}}), true
	case "set_thinking_level":
		s.setThinking(command.Level)
		return success(command.ID, "set_thinking_level", nil), true
	case "cycle_thinking_level":
		return success(command.ID, "cycle_thinking_level", map[string]any{"level": s.cycleThinking()}), true
	case "set_steering_mode":
		s.setQueueMode("steering", command.Mode)
		return success(command.ID, "set_steering_mode", nil), true
	case "set_follow_up_mode":
		s.setQueueMode("follow_up", command.Mode)
		return success(command.ID, "set_follow_up_mode", nil), true
	case "compact":
		instructions := firstNonEmpty(command.CustomInstructions, command.CustomInstructionsAlt)
		return success(command.ID, "compact", map[string]any{"compacted": false, "custom_instructions": instructions}), true
	case "set_auto_compaction":
		s.setAutoCompaction(command.Enabled)
		return success(command.ID, "set_auto_compaction", nil), true
	case "set_auto_retry":
		s.setAutoRetry(command.Enabled)
		return success(command.ID, "set_auto_retry", nil), true
	case "abort_retry":
		return success(command.ID, "abort_retry", nil), true
	case "bash":
		data, err := s.runBash(ctx, command.CommandText)
		if err != nil {
			return failure(command.ID, "bash", err.Error()), true
		}
		return success(command.ID, "bash", data), true
	case "abort_bash":
		return success(command.ID, "abort_bash", nil), true
	case "get_session_stats":
		return success(command.ID, "get_session_stats", s.sessionStats()), true
	case "export_html":
		return success(command.ID, "export_html", map[string]any{"path": command.OutputPath}), true
	case "switch_session":
		s.switchSession(command.SessionPath)
		return success(command.ID, "switch_session", map[string]any{"cancelled": false}), true
	case "fork":
		s.newSession(command.EntryID)
		return success(command.ID, "fork", map[string]any{"text": "", "cancelled": false}), true
	case "clone":
		s.newSession(s.currentSessionID())
		return success(command.ID, "clone", map[string]any{"cancelled": false}), true
	case "get_fork_messages":
		return success(command.ID, "get_fork_messages", map[string]any{"messages": s.userMessages()}), true
	case "get_entries":
		return success(command.ID, "get_entries", map[string]any{"entries": s.entriesSince(command.Since), "leafId": s.leafID()}), true
	case "get_tree":
		return success(command.ID, "get_tree", map[string]any{"tree": s.treeData(), "leafId": s.leafID()}), true
	case "get_last_assistant_text":
		return success(command.ID, "get_last_assistant_text", map[string]any{"text": s.lastAssistantText()}), true
	case "set_session_name":
		if strings.TrimSpace(command.Name) == "" {
			return failure(command.ID, "set_session_name", "session name cannot be empty"), true
		}
		s.setSessionName(command.Name)
		return success(command.ID, "set_session_name", nil), true
	case "get_messages":
		return success(command.ID, "get_messages", map[string]any{"messages": s.messagesData()}), true
	case "get_commands":
		return success(command.ID, "get_commands", map[string]any{"commands": []any{}}), true
	case "shutdown":
		s.requestShutdown()
		return success(command.ID, "shutdown", nil), true
	case "":
		return failure(command.ID, "unknown", "missing command type"), true
	default:
		return failure(command.ID, command.Type, "unknown command: "+command.Type), true
	}
}

func (s *Server) startPrompt(ctx context.Context, id string, text string) response {
	text = strings.TrimSpace(text)
	if text == "" {
		return failure(id, "prompt", "message cannot be empty")
	}

	s.mu.Lock()
	a := s.agent
	if a == nil {
		s.mu.Unlock()
		return failure(id, "prompt", "rpc prompt requires agent handler")
	}
	if s.running {
		s.mu.Unlock()
		return failure(id, "prompt", "agent busy")
	}
	// Mark busy before the goroutine starts so immediate RPC commands see the active stream.
	s.running = true
	promptText := s.mergeSteeringLocked(text)
	s.addMessageLocked("user", promptText)
	s.mu.Unlock()

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		err := a.Prompt(ctx, agent.Prompt{Text: promptText})
		s.finishPrompt(ctx, err)
	}()
	return success(id, "prompt", nil)
}

func (s *Server) finishPrompt(ctx context.Context, err error) {
	var next string
	s.mu.Lock()
	shuttingDown := s.shutdown
	if err == nil && !shuttingDown && len(s.followUpQueue) > 0 {
		// Follow-ups run only after the active provider turn has released the agent.
		next = s.popFollowUpLocked()
	}
	s.running = false
	s.mu.Unlock()

	if err != nil && !shuttingDown && !errors.Is(err, context.Canceled) {
		_ = s.write(map[string]any{"type": "prompt_error", "error": err.Error()})
	}
	if next != "" {
		_ = s.write(map[string]any{"type": "queue_update", "data": s.queueData()})
		_ = s.startPrompt(ctx, "", next)
	}
}

func (s *Server) recordEvent(ev agent.Event) {
	if ev.Type != "turn_end" {
		return
	}
	text := turnEndText(ev.Data)
	if text == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.addMessageLocked("assistant", text)
}

func (s *Server) write(value any) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	if s.writeErr != nil {
		return s.writeErr
	}
	if err := WriteLine(s.stdout, value); err != nil {
		s.writeErr = err
		return err
	}
	return nil
}

func (s *Server) setWriteErr(err error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	if s.writeErr == nil {
		s.writeErr = err
	}
}

func (s *Server) currentWriteErr() error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	return s.writeErr
}

func (s *Server) waitIdle(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()
	select {
	case <-ctx.Done():
		return fmt.Errorf("wait rpc idle: %w", ctx.Err())
	case <-done:
		return nil
	}
}

func (s *Server) abort() {
	s.mu.Lock()
	a := s.agent
	s.mu.Unlock()
	if a != nil {
		a.Abort()
	}
}

func (s *Server) isRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

func (s *Server) shouldStop() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.shutdown
}

func (s *Server) requestShutdown() {
	s.mu.Lock()
	s.shutdown = true
	s.mu.Unlock()
	s.abort()
}

func (s *Server) enqueueSteering(text string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if strings.TrimSpace(text) != "" {
		s.steeringQueue = append(s.steeringQueue, text)
	}
}

func (s *Server) enqueueFollowUp(text string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if strings.TrimSpace(text) != "" {
		s.followUpQueue = append(s.followUpQueue, text)
	}
}

func (s *Server) mergeSteeringLocked(text string) string {
	if len(s.steeringQueue) == 0 {
		return text
	}
	count := len(s.steeringQueue)
	if s.steeringMode == "one-at-a-time" {
		count = 1
	}
	steering := strings.Join(s.steeringQueue[:count], "\n")
	s.steeringQueue = append([]string(nil), s.steeringQueue[count:]...)
	return steering + "\n\n" + text
}

func (s *Server) popFollowUpLocked() string {
	if len(s.followUpQueue) == 0 {
		return ""
	}
	count := len(s.followUpQueue)
	if s.followUpMode == "one-at-a-time" {
		count = 1
	}
	next := strings.Join(s.followUpQueue[:count], "\n")
	s.followUpQueue = append([]string(nil), s.followUpQueue[count:]...)
	return next
}

func (s *Server) addMessageLocked(role string, text string) {
	s.nextMessageIndex++
	s.messages = append(s.messages, rpcMessage{
		ID:      fmt.Sprintf("m%d", s.nextMessageIndex),
		Role:    role,
		Content: text,
	})
}

func (s *Server) newSession(parent string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessionID = newID("session")
	s.messages = nil
	s.nextMessageIndex = 0
	if parent != "" {
		s.settings["parent_session"] = parent
	}
}

func (s *Server) currentSessionID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sessionID
}

func (s *Server) switchSession(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if strings.TrimSpace(path) == "" {
		s.sessionID = newID("session")
		return
	}
	s.sessionID = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
}

func (s *Server) setSessionName(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessionName = strings.TrimSpace(name)
}

func (s *Server) setSettings(settings map[string]any, persist bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for key, value := range settings {
		s.settings[key] = value
	}
	s.settings["persist"] = persist
}

func (s *Server) setModel(providerID string, modelID string) error {
	s.mu.Lock()
	providerID = firstNonEmpty(providerID, s.provider)
	api := firstNonEmpty(s.api, "chat")
	a := s.agent
	s.mu.Unlock()

	if a != nil {
		if err := a.SetModel(providerID, api, modelID); err != nil {
			return err
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.provider = providerID
	s.api = api
	s.model = modelID
	s.modelLabel = modelID
	return nil
}

func (s *Server) setThinking(level string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.thinkingLevel = firstNonEmpty(level, "off")
}

func (s *Server) cycleThinking() string {
	levels := []string{"off", "minimal", "low", "medium", "high", "xhigh"}
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, level := range levels {
		if s.thinkingLevel == level {
			s.thinkingLevel = levels[(i+1)%len(levels)]
			return s.thinkingLevel
		}
	}
	s.thinkingLevel = "off"
	return s.thinkingLevel
}

func (s *Server) setQueueMode(queue string, mode string) {
	if mode != "one-at-a-time" {
		mode = "all"
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if queue == "steering" {
		s.steeringMode = mode
		return
	}
	s.followUpMode = mode
}

func (s *Server) setAutoCompaction(enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.autoCompaction = enabled
}

func (s *Server) setAutoRetry(enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.autoRetry = enabled
}

func (s *Server) runBash(ctx context.Context, command string) (any, error) {
	args, err := json.Marshal(map[string]any{"command": command})
	if err != nil {
		return nil, fmt.Errorf("encode bash command: %w", err)
	}
	result, err := bash.Run(ctx, s.cwd, string(args), defaultBashOutputLimit)
	if err != nil {
		return nil, err
	}
	var data any
	if err := json.Unmarshal([]byte(result.Content), &data); err != nil {
		return result.Content, nil
	}
	return data, nil
}

func (s *Server) stateData() map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	return map[string]any{
		"session_id":                s.sessionID,
		"sessionId":                 s.sessionID,
		"session_name":              s.sessionName,
		"cwd":                       s.cwd,
		"model":                     s.modelDataLocked(),
		"provider":                  s.provider,
		"api":                       s.api,
		"busy":                      s.running,
		"isStreaming":               s.running,
		"isCompacting":              false,
		"thinkingLevel":             s.thinkingLevel,
		"steeringMode":              s.steeringMode,
		"followUpMode":              s.followUpMode,
		"autoCompactionEnabled":     s.autoCompaction,
		"autoRetryEnabled":          s.autoRetry,
		"messageCount":              len(s.messages),
		"pendingMessageCount":       len(s.followUpQueue),
		"steering_queue":            append([]string(nil), s.steeringQueue...),
		"follow_up_queue":           append([]string(nil), s.followUpQueue...),
		"active_leaf":               s.leafIDLocked(),
		"settings":                  cloneMap(s.settings),
		"stdout_protocol_exclusive": true,
	}
}

func (s *Server) settingsData() map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	return cloneMap(s.settings)
}

func (s *Server) modelData() map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.modelDataLocked()
}

func (s *Server) modelDataLocked() map[string]any {
	return map[string]any{
		"provider":     s.provider,
		"api":          s.api,
		"id":           s.model,
		"modelId":      s.model,
		"display_name": s.modelLabel,
		"displayName":  s.modelLabel,
	}
}

func (s *Server) thinking() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.thinkingLevel
}

func (s *Server) queueData() map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	return map[string]any{
		"steering":  append([]string(nil), s.steeringQueue...),
		"follow_up": append([]string(nil), s.followUpQueue...),
	}
}

func (s *Server) sessionStats() map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	var userCount, assistantCount int
	for _, message := range s.messages {
		switch message.Role {
		case "user":
			userCount++
		case "assistant":
			assistantCount++
		}
	}
	return map[string]any{
		"sessionId":      s.sessionID,
		"messageCount":   len(s.messages),
		"userCount":      userCount,
		"assistantCount": assistantCount,
	}
}

func (s *Server) userMessages() []map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	var messages []map[string]any
	for _, message := range s.messages {
		if message.Role == "user" {
			messages = append(messages, map[string]any{"entryId": message.ID, "text": message.Content})
		}
	}
	return messages
}

func (s *Server) entriesSince(since string) []rpcMessage {
	s.mu.Lock()
	defer s.mu.Unlock()
	messages := append([]rpcMessage(nil), s.messages...)
	if since == "" {
		return messages
	}
	for i, message := range messages {
		if message.ID == since {
			return messages[i+1:]
		}
	}
	return nil
}

func (s *Server) treeData() []map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	tree := make([]map[string]any, 0, len(s.messages))
	for i, message := range s.messages {
		parent := ""
		if i > 0 {
			parent = s.messages[i-1].ID
		}
		tree = append(tree, map[string]any{
			"id":       message.ID,
			"parentId": parent,
			"role":     message.Role,
			"text":     message.Content,
		})
	}
	return tree
}

func (s *Server) leafID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.leafIDLocked()
}

func (s *Server) leafIDLocked() string {
	if len(s.messages) == 0 {
		return ""
	}
	return s.messages[len(s.messages)-1].ID
}

func (s *Server) lastAssistantText() any {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := len(s.messages) - 1; i >= 0; i-- {
		if s.messages[i].Role == "assistant" {
			return s.messages[i].Content
		}
	}
	return nil
}

func (s *Server) messagesData() []rpcMessage {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]rpcMessage(nil), s.messages...)
}

func success(id string, command string, data any) response {
	resp := response{"type": "response", "command": command, "success": true}
	if id != "" {
		resp["id"] = id
	}
	if data != nil {
		resp["data"] = data
	}
	return resp
}

func failure(id string, command string, message string) response {
	resp := response{"type": "response", "command": command, "success": false, "error": message}
	if id != "" {
		resp["id"] = id
	}
	return resp
}

func turnEndText(data any) string {
	values, ok := data.(map[string]string)
	if ok {
		return values["text"]
	}
	generic, ok := data.(map[string]any)
	if ok {
		text, _ := generic["text"].(string)
		return text
	}
	return ""
}

func cloneMap(values map[string]any) map[string]any {
	cloned := make(map[string]any, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func newID(prefix string) string {
	var data [8]byte
	if _, err := rand.Read(data[:]); err != nil {
		return prefix + "-fallback"
	}
	return prefix + "-" + hex.EncodeToString(data[:])
}
