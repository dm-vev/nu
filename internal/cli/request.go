package cli

// CommandKind is the parsed top-level command.
type CommandKind string

const (
	CommandChat       CommandKind = "chat"
	CommandHelp       CommandKind = "help"
	CommandVersion    CommandKind = "version"
	CommandPackage    CommandKind = "package"
	CommandConfig     CommandKind = "config"
	CommandUpdate     CommandKind = "update"
	CommandExport     CommandKind = "export"
	CommandShare      CommandKind = "share"
	CommandListModels CommandKind = "list-models"
)

// Mode is the requested runtime mode.
type Mode string

const (
	ModeInteractive Mode = "interactive"
	ModePrint       Mode = "print"
	ModeJSON        Mode = "json"
	ModeRPC         Mode = "rpc"
)

// Request is the parsed CLI request passed to app mode dispatch.
type Request struct {
	Command        CommandKind
	Mode           Mode
	Prompt         []string
	FileArgs       []string
	ExtensionFlags []string
}

// Diagnostic is a non-fatal parse diagnostic.
type Diagnostic struct {
	Message string
}
