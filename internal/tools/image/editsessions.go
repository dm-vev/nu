package image

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"nu/internal/contracts"
	"nu/internal/multitenancy"
)

type editSessionEntry struct {
	session   contracts.ImageEditSession
	lastUsed  time.Time
	orgID     string
	createdAt time.Time
}

func (t *EditTool) startSession(ctx context.Context, prompt, aspectRatio, imageSize string) (string, error) {
	// Get organization ID for session tracking
	orgID, _ := multitenancy.GetOrgID(ctx)

	// Check session limits
	t.sessionsMu.RLock()
	orgSessionCount := 0
	for _, entry := range t.sessions {
		if entry.orgID == orgID {
			orgSessionCount++
		}
	}
	t.sessionsMu.RUnlock()

	if orgSessionCount >= t.maxSessions {
		return "", fmt.Errorf("maximum number of concurrent sessions (%d) reached. Please end an existing session first", t.maxSessions)
	}

	// Create session options
	sessionOpts := &contracts.ImageEditSessionOptions{
		Model: t.defaultModel,
	}

	// Create new session
	session, err := t.editor.CreateImageEditSession(ctx, sessionOpts)
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	// Generate session ID
	sessionID := uuid.New().String()

	// Store session
	t.sessionsMu.Lock()
	t.sessions[sessionID] = &editSessionEntry{
		session:   session,
		lastUsed:  time.Now(),
		orgID:     orgID,
		createdAt: time.Now(),
	}
	t.sessionsMu.Unlock()

	// If prompt provided, generate initial image
	if prompt != "" {
		if len(prompt) > t.maxPromptLen {
			// Clean up session on validation error
			t.sessionsMu.Lock()
			delete(t.sessions, sessionID)
			t.sessionsMu.Unlock()
			_ = session.Close()
			return "", fmt.Errorf("prompt exceeds maximum length of %d characters", t.maxPromptLen)
		}

		resp, err := session.SendMessage(ctx, prompt, &contracts.ImageEditOptions{
			AspectRatio: aspectRatio,
			ImageSize:   imageSize,
		})
		if err != nil {
			// Clean up session on error
			t.sessionsMu.Lock()
			delete(t.sessions, sessionID)
			t.sessionsMu.Unlock()
			_ = session.Close()
			return "", fmt.Errorf("failed to generate initial image: %w", err)
		}
		return t.formatResponse(ctx, sessionID, resp, prompt, true)
	}

	return fmt.Sprintf(`Image editing session started successfully.

Session ID: %s

Use this session ID with action='edit' to generate and modify images. The session maintains conversation context, so you can iteratively refine your images.

When done, use action='end_session' to close the session.`, sessionID), nil
}

func (t *EditTool) editImage(ctx context.Context, sessionID, prompt, aspectRatio, imageSize string) (string, error) {
	// Validate prompt length
	if len(prompt) > t.maxPromptLen {
		return "", fmt.Errorf("prompt exceeds maximum length of %d characters", t.maxPromptLen)
	}

	// Get session
	t.sessionsMu.RLock()
	entry, exists := t.sessions[sessionID]
	t.sessionsMu.RUnlock()

	if !exists {
		return "", fmt.Errorf("%w: %s", contracts.ErrSessionNotFound, sessionID)
	}

	// Check session timeout
	if time.Since(entry.lastUsed) > t.sessionTimeout {
		// Session expired, clean it up
		t.sessionsMu.Lock()
		delete(t.sessions, sessionID)
		t.sessionsMu.Unlock()
		_ = entry.session.Close()
		return "", fmt.Errorf("%w: session %s has expired after %v of inactivity", contracts.ErrSessionExpired, sessionID, t.sessionTimeout)
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

	return t.formatResponse(ctx, sessionID, resp, prompt, false)
}

func (t *EditTool) endSession(ctx context.Context, sessionID string) (string, error) {
	t.sessionsMu.Lock()
	defer t.sessionsMu.Unlock()

	entry, exists := t.sessions[sessionID]
	if !exists {
		return "", fmt.Errorf("%w: %s", contracts.ErrSessionNotFound, sessionID)
	}

	// Get history count before closing
	historyLen := len(entry.session.GetHistory())
	duration := time.Since(entry.createdAt)

	// Close and remove session
	_ = entry.session.Close()
	delete(t.sessions, sessionID)

	return fmt.Sprintf(`Session %s closed successfully.

Session duration: %v
Total turns: %d

The session context has been cleared.`, sessionID, duration.Round(time.Second), historyLen/2), nil
}

// cleanupExpiredSessions runs periodically to clean up expired sessions
func (t *EditTool) cleanupExpiredSessions() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		t.sessionsMu.Lock()
		now := time.Now()
		for sessionID, entry := range t.sessions {
			if now.Sub(entry.lastUsed) > t.sessionTimeout {
				_ = entry.session.Close()
				delete(t.sessions, sessionID)
			}
		}
		t.sessionsMu.Unlock()
	}
}

// GetActiveSessions returns the number of active sessions (useful for monitoring)
func (t *EditTool) GetActiveSessions() int {
	t.sessionsMu.RLock()
	defer t.sessionsMu.RUnlock()
	return len(t.sessions)
}

// GetActiveSessionsForOrg returns the number of active sessions for a specific organization
func (t *EditTool) GetActiveSessionsForOrg(orgID string) int {
	t.sessionsMu.RLock()
	defer t.sessionsMu.RUnlock()

	count := 0
	for _, entry := range t.sessions {
		if entry.orgID == orgID {
			count++
		}
	}
	return count
}
