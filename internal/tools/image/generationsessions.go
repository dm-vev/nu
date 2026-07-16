package image

import (
	"context"
	"fmt"
	"time"

	"nu/internal/contracts"
	"nu/internal/multitenancy"
)

// generationSessionEntry tracks an active multi-turn editing session.
type generationSessionEntry struct {
	session   contracts.ImageEditSession
	lastUsed  time.Time
	orgID     string
	createdAt time.Time
}

// executeMultiTurn handles multi-turn image editing with automatic session management
func (t *GenerationTool) executeMultiTurn(ctx context.Context, action, prompt, aspectRatio, imageSize string) (string, error) {
	// Get session key from context (org + thread)
	sessionKey := t.getSessionKey(ctx)

	switch action {
	case "generate":
		// For "generate", we start a new session (closing any existing one)
		return t.generateWithSession(ctx, sessionKey, prompt, aspectRatio, imageSize)

	case "edit":
		// For "edit", we continue the existing session
		return t.editInSession(ctx, sessionKey, prompt, aspectRatio, imageSize)

	case "end_session":
		// Close the session
		return t.endSession(ctx, sessionKey)

	default:
		return "", fmt.Errorf("unknown action: %s. Valid actions: generate, edit, end_session", action)
	}
}

// getSessionKey returns a unique key for the current context.
// Uses org ID if available, otherwise defaults to "default".
// Note: In a multi-thread scenario, you may want to extend this
// to include thread information from the context.
func (t *GenerationTool) getSessionKey(ctx context.Context) string {
	orgID, _ := multitenancy.GetOrgID(ctx)

	if orgID == "" {
		orgID = "default"
	}

	return orgID
}

// generateWithSession creates a new session and generates an initial image
func (t *GenerationTool) generateWithSession(ctx context.Context, sessionKey, prompt, aspectRatio, imageSize string) (string, error) {
	if prompt == "" {
		return "", fmt.Errorf("prompt is required for generating an image")
	}

	if len(prompt) > t.maxPromptLen {
		return "", fmt.Errorf("prompt exceeds maximum length of %d characters", t.maxPromptLen)
	}

	// Close any existing session for this key
	t.sessionsMu.Lock()
	if existing, ok := t.sessions[sessionKey]; ok {
		_ = existing.session.Close()
		delete(t.sessions, sessionKey)
	}
	t.sessionsMu.Unlock()

	// Check session limits
	orgID, _ := multitenancy.GetOrgID(ctx)
	if t.maxSessionsPerOrg > 0 {
		t.sessionsMu.RLock()
		orgCount := 0
		for _, entry := range t.sessions {
			if entry.orgID == orgID {
				orgCount++
			}
		}
		t.sessionsMu.RUnlock()

		if orgCount >= t.maxSessionsPerOrg {
			return "", fmt.Errorf("maximum number of concurrent sessions (%d) reached", t.maxSessionsPerOrg)
		}
	}

	// Create new session
	session, err := t.multiTurnEditor.CreateImageEditSession(ctx, &contracts.ImageEditSessionOptions{
		Model: t.multiTurnModel,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create editing session: %w", err)
	}

	// Store session
	t.sessionsMu.Lock()
	t.sessions[sessionKey] = &generationSessionEntry{
		session:   session,
		lastUsed:  time.Now(),
		orgID:     orgID,
		createdAt: time.Now(),
	}
	t.sessionsMu.Unlock()

	// Generate initial image
	if imageSize == "" {
		imageSize = "1K"
	}

	resp, err := session.SendMessage(ctx, prompt, &contracts.ImageEditOptions{
		AspectRatio: aspectRatio,
		ImageSize:   imageSize,
	})
	if err != nil {
		// Clean up session on error
		t.sessionsMu.Lock()
		delete(t.sessions, sessionKey)
		t.sessionsMu.Unlock()
		_ = session.Close()
		return "", fmt.Errorf("failed to generate initial image: %w", err)
	}

	return t.formatMultiTurnResponse(ctx, resp, prompt, true)
}

// editInSession modifies the current image in an existing session
func (t *GenerationTool) editInSession(ctx context.Context, sessionKey, prompt, aspectRatio, imageSize string) (string, error) {
	if prompt == "" {
		return "", fmt.Errorf("prompt is required for editing")
	}

	if len(prompt) > t.maxPromptLen {
		return "", fmt.Errorf("prompt exceeds maximum length of %d characters", t.maxPromptLen)
	}

	// Get session
	t.sessionsMu.RLock()
	entry, exists := t.sessions[sessionKey]
	t.sessionsMu.RUnlock()

	if !exists {
		// No active session - start one automatically
		return t.generateWithSession(ctx, sessionKey, prompt, aspectRatio, imageSize)
	}

	// Check session timeout
	if time.Since(entry.lastUsed) > t.sessionTimeout {
		// Session expired, clean it up and start fresh
		t.sessionsMu.Lock()
		delete(t.sessions, sessionKey)
		t.sessionsMu.Unlock()
		_ = entry.session.Close()
		return t.generateWithSession(ctx, sessionKey, prompt, aspectRatio, imageSize)
	}

	// Update last used time
	entry.lastUsed = time.Now()

	// Send edit request
	resp, err := entry.session.SendMessage(ctx, prompt, &contracts.ImageEditOptions{
		AspectRatio: aspectRatio,
		ImageSize:   imageSize,
	})
	if err != nil {
		return "", fmt.Errorf("failed to edit image: %w", err)
	}

	return t.formatMultiTurnResponse(ctx, resp, prompt, false)
}

// endSession closes the current editing session
func (t *GenerationTool) endSession(ctx context.Context, sessionKey string) (string, error) {
	t.sessionsMu.Lock()
	defer t.sessionsMu.Unlock()

	entry, exists := t.sessions[sessionKey]
	if !exists {
		return "No active editing session to close.", nil
	}

	// Get stats before closing
	historyLen := len(entry.session.GetHistory())
	duration := time.Since(entry.createdAt)

	// Close and remove session
	_ = entry.session.Close()
	delete(t.sessions, sessionKey)

	return fmt.Sprintf("Editing session closed.\n\nSession duration: %v\nTotal turns: %d",
		duration.Round(time.Second), historyLen/2), nil
}

// cleanupExpiredSessions runs periodically to clean up expired sessions
func (t *GenerationTool) cleanupExpiredSessions() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		t.sessionsMu.Lock()
		now := time.Now()
		for sessionKey, entry := range t.sessions {
			if now.Sub(entry.lastUsed) > t.sessionTimeout {
				_ = entry.session.Close()
				delete(t.sessions, sessionKey)
			}
		}
		t.sessionsMu.Unlock()
	}
}

// GetActiveSessions returns the number of active sessions (useful for monitoring)
func (t *GenerationTool) GetActiveSessions() int {
	if !t.multiTurnEnabled {
		return 0
	}
	t.sessionsMu.RLock()
	defer t.sessionsMu.RUnlock()
	return len(t.sessions)
}
