package rpc

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/dm-vev/nu/internal/agentui"
)

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
	agent            *agentui.Agent
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
func (s *Server) SetAgent(a *agentui.Agent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.agent = a
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
