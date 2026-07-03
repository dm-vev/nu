# Functional Spec

Each item lists required behavior and the first tests to write. More tests can
be added without changing the spec.

## CLI And Modes

### NUF-001 CLI Parse

Nu accepts Pi-compatible common flags: `--help`, `--version`, `--mode`,
`--print`, `--provider`, `--model`, `--api-key`, `--thinking`, `--continue`,
`--resume`, `--session`, `--session-id`, `--fork`, `--session-dir`,
`--no-session`, `--name`, `--models`, `--tools`, `--exclude-tools`,
`--no-tools`, `--no-builtin-tools`, `--extension`, `--no-extensions`,
`--skill`, `--no-skills`, `--prompt-template`, `--no-prompt-templates`,
`--theme`, `--no-themes`, `--no-context-files`, `--list-models`, `--offline`,
`--verbose`, `--approve`, and `--no-approve`.

Tests:

- `TestNUF001ParseKnownFlags`
- `TestNUF001UnknownFlagsArePreservedForExtensions`
- `TestNUF001InvalidThinkingLevelReportsDiagnostic`

### NUF-002 Modes

Nu supports interactive default mode, print mode, JSON event stream mode, RPC
mode, export mode, share command, package commands, config mode, and update
mode.

Tests:

- `TestNUF002DispatchPrintMode`
- `TestNUF002DispatchRPCMode`
- `TestNUF002DispatchPackageCommand`
- `TestNUF002DispatchShareCommand`

## Config, Trust, Auth

### NUF-010 Settings Resolution

Global settings load from `~/.nu/agent/settings.json`; project settings load
from `.nu/settings.json` only when trusted or explicitly approved. CLI flags
override settings.

Tests:

- `TestNUF010ProjectSettingsIgnoredWithoutTrust`
- `TestNUF010CLIOverridesSettings`

### NUF-011 Project Trust

Interactive mode may ask for trust. Non-interactive modes use configured default
trust plus `--approve` or `--no-approve`.

Tests:

- `TestNUF011ApproveAllowsProjectResources`
- `TestNUF011NoApproveBlocksProjectResources`

### NUF-020 Auth Storage

API keys and OAuth tokens are stored in `auth.json`. Auth file values override
process environment. Env interpolation and command interpolation are supported
for configured values. Bedrock auth requires both AWS access key id and secret
access key before it is treated as available.

Tests:

- `TestNUF020AuthFileBeatsEnvironment`
- `TestNUF020EnvInterpolation`
- `TestNUF020CommandInterpolation`
- `TestBedrockEnvRequiresAccessKeyAndSecret`

## Providers And Models

### NUF-030 Provider APIs

Nu supports provider adapters for OpenAI Chat Completions, OpenAI Responses,
Anthropic Messages, Google Generative AI, AWS Bedrock, and OpenAI-compatible
custom providers.

Tool-use continuations preserve the assistant tool-call message before tool
result messages so provider-specific APIs can build valid follow-up payloads.

Tests:

- `TestNUF030OpenAIChatRequestShape`
- `TestNUF030OpenAIResponsesToolCallStream`
- `TestOpenAIChatPayloadIncludesAssistantToolCalls`
- `TestOpenAIResponsesPayloadIncludesFunctionCallHistory`
- `TestNUF030AnthropicMessagesRequestShape`
- `TestAnthropicPayloadIncludesAssistantToolUse`
- `TestGoogleGenerateContentRequestShape`
- `TestGooglePayloadIncludesFunctionCallAndResponse`
- `TestBedrockConverseRequestShape`
- `TestBedrockPayloadIncludesToolUseAndResult`
- `TestBedrockSignAddsAuthorization`
- `TestBedrockRejectsOversizedEventFrame`

### NUF-031 Model Registry

Built-in models, custom `models.json`, provider auth state, model pattern
matching, enabled model cycling, input capability, context window, max output,
cost, configurable display names, and thinking support are represented in a
registry.

Tests:

- `TestNUF031ModelPatternSelectsProviderAndModel`
- `TestNUF031UnavailableModelsHiddenWithoutAuth`
- `TestNUF031CustomModelsOverrideBuiltins`
- `TestNUF031CustomModelDisplayNameLoads`
- `TestCustomModelsCanDisableEntry`

### NUF-032 Thinking Levels

Thinking levels are `off`, `minimal`, `low`, `medium`, `high`, and `xhigh`.
Provider adapters translate these into provider-specific request fields.

Tests:

- `TestNUF032ThinkingLevelMapping`
- `TestNUF032UnsupportedThinkingLevelFallsBackOrErrors`

## Agent Core

### NUF-050 Agent Loop

The agent loops over provider responses and tool results until the assistant
stops without tool calls, the user aborts, or retry policy is exhausted.

When a provider requests a tool, the next provider request includes both the
assistant tool-call message and the tool result message in order.

Tests:

- `TestNUF050TextOnlyTurnEnds`
- `TestNUF050ToolCallFeedsResultBackToProvider`
- `TestNUF050AbortStopsProviderAndTools`

### NUF-051 Tool Execution Policy

Tools run in parallel by default unless a tool or config asks for sequential
execution. Tool calls are validated before execution.

Tests:

- `TestNUF051ParallelToolCallsPreserveResultOrder`
- `TestNUF051SequentialToolRunsInOrder`
- `TestNUF051InvalidArgsReturnToolError`

### NUF-052 Queues

Interactive and RPC modes support steering messages and follow-up messages.
Steering can be delivered one-at-a-time or all at once. Follow-ups run after the
agent becomes idle.

Tests:

- `TestNUF052SteeringDeliveredBeforeNextProviderCall`
- `TestNUF052FollowUpWaitsForIdle`

### NUF-053 Retry

Transient provider errors use configurable exponential backoff. Provider-level
retry delays above the configured cap fail fast.

Tests:

- `TestNUF053RetriesTransientError`
- `TestNUF053LongRetryAfterFailsFast`

## Messages And Events

### NUF-060 Message Types

Nu supports user, assistant, tool result, bash execution, custom, branch summary,
and compaction summary messages with text, image, thinking, and tool call
content blocks.

Tests:

- `TestNUF060MessageJSONRoundTrip`
- `TestNUF060ImageContentRoundTrip`

### NUF-061 Event Stream

Nu emits lifecycle, turn, message, tool execution, queue, compaction, retry, and
session events. JSON mode writes one event per line.

Tests:

- `TestNUF061JSONModeEmitsSessionThenEvents`
- `TestNUF061ToolEventsWrapExecution`

## Built-in Tools

### NUF-070 Read

Reads text files with offset/limit and truncation. Reads supported image files
as image attachments unless blocked.

Tests:

- `TestNUF070ReadTextWithOffsetLimit`
- `TestNUF070ReadTruncatesLargeFile`
- `TestNUF070ReadImageAttachment`

### NUF-071 Write

Creates or overwrites files relative to cwd, respecting cancellation and
serialized mutations.

Tests:

- `TestNUF071WriteCreatesFile`
- `TestNUF071ConcurrentWritesSamePathSerialize`

### NUF-072 Edit

Applies one or more exact text replacements against original file contents,
preserves line endings, rejects missing or ambiguous matches, and returns a
unified patch.

Tests:

- `TestNUF072EditSingleReplacement`
- `TestNUF072EditRejectsAmbiguousOldText`
- `TestNUF072EditPreservesCRLF`

### NUF-073 Bash

Runs shell commands in cwd with optional timeout, streams updates, captures
stdout/stderr, returns exit code, truncates display output, and persists full
truncated output to a temp file.

Tests:

- `TestNUF073BashCapturesStdoutAndStderr`
- `TestNUF073BashTimeoutKillsProcess`
- `TestNUF073BashTruncatesAndPersistsFullOutput`

### NUF-074 Grep

Searches file contents with regex or literal matching, optional glob,
ignore-case, context lines, gitignore support, limits, and long-line truncation.

Tests:

- `TestNUF074GrepLiteralAndRegex`
- `TestNUF074GrepRespectsGitignore`

### NUF-075 Find

Finds files by glob under a root path, respects gitignore, returns sorted
relative paths, and enforces limits.

Tests:

- `TestNUF075FindGlob`
- `TestNUF075FindRespectsGitignore`

### NUF-076 Ls

Lists directory entries sorted alphabetically, includes dotfiles, marks
directories with `/`, and truncates large output.

Tests:

- `TestNUF076LsSortedWithDirs`
- `TestNUF076LsRejectsNonDirectory`

## Sessions

### NUF-080 Session Storage

Sessions are append-only JSONL trees. Entries include schema version, session id,
cwd, parent id, timestamp, type, and payload.

Tests:

- `TestNUF080SessionAppendBuildsTree`
- `TestNUF080SessionLoadRejectsBrokenParent`
- `TestSessionExportImportRoundTrip`
- `TestSessionAppendUsesRefCWD`
- `TestSessionImportRejectsOversizedInput`

### NUF-081 Resume, Fork, Clone

Nu can continue the latest session for cwd, resume by path or partial id, fork
from an earlier user message, and clone the active branch to a new session.

Tests:

- `TestNUF081ContinueLatestByCWD`
- `TestNUF081ResumeByPathOrPartialID`
- `TestNUF081ForkStartsNewFileFromUserEntry`
- `TestNUF081CloneCopiesActiveBranch`

### NUF-082 Tree Navigation

Interactive tree navigation can move the active leaf, edit/resubmit user
messages to create branches, label entries, filter views, and optionally create
branch summaries.

Tests:

- `TestNUF082SelectingUserEntryPreloadsEditor`
- `TestNUF082SelectingAssistantEntryMovesLeaf`
- `TestSessionStateEntrySetsActiveLeaf`

## Compaction

### NUF-090 Auto Compaction

When estimated context exceeds `contextWindow - reserveTokens`, Nu summarizes
old context while keeping recent messages and valid tool-call boundaries.

Tests:

- `TestNUF090CompactionKeepsRecentBudget`
- `TestCompactionCompactsOversizedSingleEntry`
- `TestNUF090CompactionDoesNotCutBeforeToolResult`

### NUF-091 Branch Summaries

When leaving a branch, Nu can summarize the abandoned branch and attach the
summary to the new position.

Tests:

- `TestNUF091BranchSummaryFindsCommonAncestor`
- `TestNUF091BranchSummaryTracksFiles`

## TUI

### NUF-100 Terminal UI

Interactive mode provides a Pi-like terminal surface: startup header, compact
keybinding help, context block, message history, bordered editor, cwd/context
footer, status, tool rendering, thinking rendering, images where supported,
selectors, overlays, and resize handling.

The first Go UI implementation must expose the same wiring boundaries as Pi:
terminal input becomes editor/command actions, agent events update a render
state, and rendering produces deterministic frames that the terminal driver
writes in place. Raw terminal integration can stay narrow, but renderer, input
decoder, editor buffer, overlay focus, and app-mode wiring are required before
further interactive features build on them.

Tests:

- `TestNUF100RendererDoesNotOverflowWidth`
- `TestNUF100ResizeInvalidatesLayout`
- `TestTerminalDrawRepaintsWithANSI`
- temporary `/tmp/nu-tui-bytecheck/check-byte-ui.py` byte harness during TUI
  implementation only; it must not be committed.
- `TestNUF002DispatchInteractiveMode`

### NUF-101 Editor

The editor supports multiline input, cursor movement, word movement, deletion,
undo, kill ring, autocomplete, external editor, pasted images, and path
completion.

The initial editor implementation must keep buffer mutations separate from
rendering, preserve cursor position through undo, submit text without mutating
history, and leave extension/autocomplete hooks as explicit state boundaries.

Tests:

- `TestNUF101EditorInsertDeleteUndo`
- `TestNUF101AutocompleteAtFileReference`

### NUF-102 Keybindings

All shortcuts are configurable by action id. Defaults can be migrated when old
ids exist.

Tests:

- `TestNUF102KeybindingConfigOverridesDefault`
- `TestNUF102OldKeybindingIDsMigrate`

### NUF-103 Themes

Nu supports built-in dark/light themes and custom JSON themes from global,
project, and package resources.

Tests:

- `TestNUF103ThemeLoadsFromResource`
- `TestNUF103BrokenThemeReportsDiagnostic`

## Slash Commands

### NUF-110 Built-in Commands

Nu supports `/login`, `/logout`, `/model`, `/scoped-models`, `/settings`,
`/resume`, `/new`, `/name`, `/session`, `/tree`, `/trust`, `/fork`, `/clone`,
`/compact`, `/copy`, `/export`, `/import`, `/share`, `/reload`, `/hotkeys`,
`/changelog`, and `/quit`.

Tests:

- `TestNUF110SlashCommandDispatch`
- `TestNUF110UnknownCommandReportsError`

## Resources

### NUF-120 Context Files

Nu loads `AGENTS.md`, `CLAUDE.md`, `.nu/SYSTEM.md`,
`~/.nu/agent/SYSTEM.md`, and append-system files according to precedence and
trust.

Tests:

- `TestNUF120ContextFilesWalkParents`
- `TestNUF120SystemPromptReplacementAndAppend`

### NUF-130 Skills

Nu discovers skills from global, project, package, settings, and CLI locations;
uses progressive disclosure; and exposes `/skill:name` commands.

Tests:

- `TestNUF130SkillDiscovery`
- `TestNUF130SkillCommandExpandsContent`

### NUF-140 Prompt Templates

Nu loads Markdown prompt templates, parses frontmatter, supports positional
arguments, defaults, and slices, and exposes templates as slash commands.

Tests:

- `TestNUF140PromptTemplateExpansion`
- `TestNUF140PromptTemplateDefaultArgument`

### NUF-150 Packages

Nu installs, lists, removes, updates, enables, and disables resource packages
from local paths, git sources, and package archives. Lifecycle scripts are never
run silently.

Tests:

- `TestNUF150LocalPackageDiscovery`
- `TestNUF150GitPackagePinnedRef`
- `TestNUF150PackageFilterIncludesAndExcludes`
- `TestNUF150PackageEnableDisableUpdatesSettings`

### NUF-160 Extensions

Nu loads trusted process extensions, handles lifecycle events, lets extensions
register tools, commands, keybindings, flags, renderers, and UI requests, and
persists extension state through session entries.

Tests:

- `TestNUF160ExtensionRegistersTool`
- `TestNUF160ExtensionCanBlockToolCall`
- `TestNUF160ExtensionShutdownRunsOnce`

## Headless Protocols

### NUF-170 JSON Mode

JSON mode writes the session header followed by event JSONL to stdout. Human
diagnostics go to stderr.

Tests:

- `TestNUF170JSONModeStdoutIsOnlyJSONL`

### NUF-171 RPC Mode

RPC mode reads JSONL commands from stdin and writes JSONL responses/events to
stdout. Stdout is protocol-only and stderr is diagnostics-only.

Nu must accept Pi's headless command names: `prompt`, `steer`, `follow_up`,
`abort`, `new_session`, `get_state`, `state`, `set_settings`, `set_model`,
`cycle_model`, `get_available_models`, `set_thinking_level`,
`cycle_thinking_level`, `set_steering_mode`, `set_follow_up_mode`, `compact`,
`set_auto_compaction`, `set_auto_retry`, `abort_retry`, `bash`, `abort_bash`,
`get_session_stats`, `export_html`, `switch_session`, `fork`, `clone`,
`get_fork_messages`, `get_entries`, `get_tree`, `get_last_assistant_text`,
`set_session_name`, `get_messages`, `get_commands`, and `shutdown`.

Commands backed by not-yet-complete Nu subsystems must still have stable RPC
wiring: validate input, update in-memory runtime state when possible, return a
structured success/error response, and never print human text to stdout.

Tests:

- `TestNUF171RPCPromptResponseCorrelation`
- `TestNUF171RPCRejectsPromptDuringStreamWithoutBehavior`
- `TestNUF171RPCShutdownWritesFinalResponse`
- `TestNUF002DispatchRPCMode`

## Export, Share, Update

### NUF-180 Export

Nu exports sessions to JSONL and standalone HTML with rendered messages, tool
calls, tool outputs, thinking, and images.

Tests:

- `TestNUF180ExportJSONLRoundTrip`
- `TestNUF180ExportHTMLContainsEscapedContent`

### NUF-181 Share

Nu can upload a private share artifact only after explicit user action and with
clear target information.

Tests:

- `TestNUF181ShareRequiresExplicitTarget`
- `TestNUF181ShareRequiresExplicitAction`

### NUF-182 Update

Nu can check and update itself and installed packages. Offline mode disables
network update checks.

Tests:

- `TestNUF182OfflineSkipsNetworkChecks`
- `TestNUF182UpdateLeavesInstallUntouchedOnFailure`
