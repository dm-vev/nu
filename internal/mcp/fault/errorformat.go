package fault

import "fmt"

// FormatUserFriendlyError creates a user-friendly error message
func FormatUserFriendlyError(err error) string {
	mcpErr, ok := err.(*Error)
	if !ok {
		return err.Error()
	}
	switch mcpErr.ErrorType {
	case ErrorTypeConnection:
		if mcpErr.ServerType == "stdio" {
			return fmt.Sprintf("Could not start MCP server '%s'. Please check that the command is installed and accessible.", mcpErr.ServerName)
		}
		return fmt.Sprintf("Could not connect to MCP server '%s'. Please check the server URL and network connectivity.", mcpErr.ServerName)
	case ErrorTypeTimeout:
		return fmt.Sprintf("MCP server '%s' took too long to respond. This might be a temporary issue - please try again.", mcpErr.ServerName)
	case ErrorTypeAuthentication:
		return fmt.Sprintf("Authentication failed for MCP server '%s'. Please check your credentials or API key.", mcpErr.ServerName)
	case ErrorTypeServerNotFound:
		return "MCP server command not found. Please ensure the server is installed correctly."
	case ErrorTypeToolNotFound:
		toolName := mcpErr.Metadata["tool_name"]
		if toolName != "" {
			return fmt.Sprintf("Tool '%s' is not available on MCP server '%s'. Try listing available tools first.", toolName, mcpErr.ServerName)
		}
		return fmt.Sprintf("Requested tool is not available on MCP server '%s'.", mcpErr.ServerName)
	case ErrorTypeToolInvalidArgs:
		return "Invalid arguments provided to MCP tool. Please check the tool's parameter requirements."
	case ErrorTypeConfiguration:
		return fmt.Sprintf("MCP configuration error: %v", mcpErr.Cause)
	case ErrorTypeServerStartup:
		return fmt.Sprintf("MCP server '%s' failed to start. Please check the server configuration.", mcpErr.ServerName)
	case ErrorTypeServerCrash:
		return fmt.Sprintf("MCP server '%s' crashed unexpectedly. Please try restarting the server.", mcpErr.ServerName)
	case ErrorTypeToolExecution:
		return fmt.Sprintf("Tool execution failed on MCP server '%s': %v", mcpErr.ServerName, mcpErr.Cause)
	case ErrorTypeProtocol:
		return fmt.Sprintf("Protocol error with MCP server '%s': %v", mcpErr.ServerName, mcpErr.Cause)
	case ErrorTypeSerialization:
		return fmt.Sprintf("Serialization error with MCP server '%s': %v", mcpErr.ServerName, mcpErr.Cause)
	case ErrorTypeValidation:
		return fmt.Sprintf("Validation error with MCP server '%s': %v", mcpErr.ServerName, mcpErr.Cause)
	case ErrorTypeUnknown:
		return fmt.Sprintf("MCP server '%s' error: %v", mcpErr.ServerName, mcpErr.Cause)
	default:
		return fmt.Sprintf("MCP server '%s' error: %v", mcpErr.ServerName, mcpErr.Cause)
	}
}
