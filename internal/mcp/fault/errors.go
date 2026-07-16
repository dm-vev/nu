package fault

import (
	"fmt"
	"strings"
	"sync"
)

// Error represents a structured error from MCP operations
type Error struct {
	Operation  string            // The operation that failed (e.g., "ListTools", "CallTool")
	ServerName string            // Name of the MCP server
	ServerType string            // Type of server (stdio, http, etc.)
	ErrorType  ErrorType         // Category of error
	Cause      error             // The underlying error
	Retryable  bool              // Whether the error might succeed on retry
	Metadata   map[string]string // Additional context
	mutex      sync.RWMutex      // Protects metadata access
}

// ErrorType categorizes different types of MCP errors
type ErrorType string

const (
	// Connection errors
	ErrorTypeConnection     ErrorType = "CONNECTION_ERROR"
	ErrorTypeTimeout        ErrorType = "TIMEOUT_ERROR"
	ErrorTypeAuthentication ErrorType = "AUTHENTICATION_ERROR"

	// Server errors
	ErrorTypeServerNotFound ErrorType = "SERVER_NOT_FOUND"
	ErrorTypeServerStartup  ErrorType = "SERVER_STARTUP_ERROR"
	ErrorTypeServerCrash    ErrorType = "SERVER_CRASH"

	// Tool errors
	ErrorTypeToolNotFound    ErrorType = "TOOL_NOT_FOUND"
	ErrorTypeToolInvalidArgs ErrorType = "TOOL_INVALID_ARGS"
	ErrorTypeToolExecution   ErrorType = "TOOL_EXECUTION_ERROR"

	// Protocol errors
	ErrorTypeProtocol      ErrorType = "PROTOCOL_ERROR"
	ErrorTypeSerialization ErrorType = "SERIALIZATION_ERROR"

	// Configuration errors
	ErrorTypeConfiguration ErrorType = "CONFIGURATION_ERROR"
	ErrorTypeValidation    ErrorType = "VALIDATION_ERROR"

	// Unknown errors
	ErrorTypeUnknown ErrorType = "UNKNOWN_ERROR"
)

// Error implements the error interface
func (e *Error) Error() string {
	var parts []string

	if e.ServerName != "" {
		parts = append(parts, fmt.Sprintf("MCP server '%s'", e.ServerName))
	} else {
		parts = append(parts, "MCP operation")
	}

	if e.Operation != "" {
		parts = append(parts, fmt.Sprintf("operation '%s'", e.Operation))
	}

	parts = append(parts, "failed")

	if e.ErrorType != ErrorTypeUnknown {
		parts = append(parts, fmt.Sprintf("(%s)", e.ErrorType))
	}

	message := strings.Join(parts, " ")

	if e.Cause != nil {
		message += fmt.Sprintf(": %v", e.Cause)
	}

	return message
}

// Unwrap returns the underlying error
func (e *Error) Unwrap() error {
	return e.Cause
}

// IsRetryable returns whether this error might succeed on retry
func (e *Error) IsRetryable() bool {
	return e.Retryable
}

// WithMetadata adds metadata to the error (thread-safe)
func (e *Error) WithMetadata(key, value string) *Error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if e.Metadata == nil {
		e.Metadata = make(map[string]string)
	}
	e.Metadata[key] = value
	return e
}

// NewError creates a new MCP error
func NewError(operation, serverName, serverType string, errorType ErrorType, cause error) *Error {
	return &Error{
		Operation:  operation,
		ServerName: serverName,
		ServerType: serverType,
		ErrorType:  errorType,
		Cause:      cause,
		Retryable:  isRetryableErrorType(errorType),
		Metadata:   make(map[string]string),
	}
}

// NewConnectionError creates a connection-related error
func NewConnectionError(serverName, serverType string, cause error) *Error {
	return NewError("Connect", serverName, serverType, ErrorTypeConnection, cause)
}

// NewTimeoutError creates a timeout error
func NewTimeoutError(operation, serverName, serverType string, cause error) *Error {
	return NewError(operation, serverName, serverType, ErrorTypeTimeout, cause)
}

// NewToolError creates a tool-related error
func NewToolError(toolName, serverName, serverType string, errorType ErrorType, cause error) *Error {
	err := NewError("CallTool", serverName, serverType, errorType, cause)
	_ = err.WithMetadata("tool_name", toolName)
	return err
}

// NewServerError creates a server-related error
func NewServerError(serverName, serverType string, errorType ErrorType, cause error) *Error {
	return NewError("ServerOperation", serverName, serverType, errorType, cause)
}

// NewConfigurationError creates a configuration error
func NewConfigurationError(operation string, cause error) *Error {
	return NewError(operation, "", "", ErrorTypeConfiguration, cause)
}

// isRetryableErrorType determines if an error type is retryable
func isRetryableErrorType(errorType ErrorType) bool {
	switch errorType {
	case ErrorTypeConnection,
		ErrorTypeTimeout,
		ErrorTypeServerStartup,
		ErrorTypeServerCrash:
		return true
	case ErrorTypeAuthentication,
		ErrorTypeServerNotFound,
		ErrorTypeToolNotFound,
		ErrorTypeToolInvalidArgs,
		ErrorTypeProtocol,
		ErrorTypeSerialization,
		ErrorTypeConfiguration,
		ErrorTypeValidation,
		ErrorTypeToolExecution,
		ErrorTypeUnknown:
		return false
	default:
		return false
	}
}

// ClassifyError attempts to classify an error based on its message
func ClassifyError(err error, operation, serverName, serverType string) *Error {
	if err == nil {
		return nil
	}

	// If it's already an Error, return it
	if mcpErr, ok := err.(*Error); ok {
		return mcpErr
	}

	errMsg := strings.ToLower(err.Error())

	// Classify based on error message content
	var errorType ErrorType
	switch {
	case strings.Contains(errMsg, "connection refused") ||
		strings.Contains(errMsg, "connection reset") ||
		strings.Contains(errMsg, "connection timeout") ||
		strings.Contains(errMsg, "no route to host") ||
		strings.Contains(errMsg, "network unreachable"):
		errorType = ErrorTypeConnection

	case strings.Contains(errMsg, "timeout") ||
		strings.Contains(errMsg, "deadline exceeded") ||
		strings.Contains(errMsg, "context deadline exceeded"):
		errorType = ErrorTypeTimeout

	case strings.Contains(errMsg, "authentication") ||
		strings.Contains(errMsg, "unauthorized") ||
		strings.Contains(errMsg, "forbidden") ||
		strings.Contains(errMsg, "invalid token"):
		errorType = ErrorTypeAuthentication

	case strings.Contains(errMsg, "not found") ||
		strings.Contains(errMsg, "no such file"):
		if operation == "CallTool" {
			errorType = ErrorTypeToolNotFound
		} else {
			errorType = ErrorTypeServerNotFound
		}

	case strings.Contains(errMsg, "invalid argument") ||
		strings.Contains(errMsg, "invalid parameter") ||
		strings.Contains(errMsg, "bad request"):
		errorType = ErrorTypeToolInvalidArgs

	case strings.Contains(errMsg, "json") ||
		strings.Contains(errMsg, "unmarshal") ||
		strings.Contains(errMsg, "marshal") ||
		strings.Contains(errMsg, "parse"):
		errorType = ErrorTypeSerialization

	case strings.Contains(errMsg, "config") ||
		strings.Contains(errMsg, "validation"):
		errorType = ErrorTypeConfiguration

	// Server crash errors - check these before protocol errors to avoid conflicts
	case strings.Contains(errMsg, "server crashed") ||
		strings.Contains(errMsg, "process exited") ||
		strings.Contains(errMsg, "broken pipe"):
		errorType = ErrorTypeServerCrash

	case strings.Contains(errMsg, "protocol") ||
		strings.Contains(errMsg, "invalid response") ||
		(strings.Contains(errMsg, "unexpected") && strings.Contains(errMsg, "response")):
		errorType = ErrorTypeProtocol

	default:
		errorType = ErrorTypeUnknown
	}

	return NewError(operation, serverName, serverType, errorType, err)
}
