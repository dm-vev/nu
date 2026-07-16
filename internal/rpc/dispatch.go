package rpc

import (
	"context"
	"strings"
)

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
