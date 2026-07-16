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
Provider reasoning/thinking stream deltas are preserved as structured provider
events instead of being merged into final assistant text.
OpenAI-compatible requests include built-in tool definitions when tools are
enabled so hosted models can actually choose tool calls.

Tests:

- `TestNUF030OpenAIChatRequestShape`
- `TestNUF030OpenAIResponsesToolCallStream`
- `TestOpenAIChatPayloadIncludesAssistantToolCalls`
- `TestOpenAIChatPayloadIncludesToolDefinitions`
- `TestOpenAIResponsesPayloadIncludesFunctionCallHistory`
- `TestOpenAIResponsesPayloadIncludesToolDefinitions`
- `TestNUF030AnthropicMessagesRequestShape`
- `TestAnthropicPayloadIncludesAssistantToolUse`
- `TestAnthropicThinkingDeltaIsPreserved`
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
Tool execution events include raw arguments on start and raw result/error state
on end so TUI/RPC clients can render command and patch blocks.
Provider requests include tool definitions from the active tool registry.

Tests:

- `TestNUF050TextOnlyTurnEnds`
- `TestNUF050ProviderRequestIncludesToolDefinitions`
- `TestNUF050ToolCallFeedsResultBackToProvider`
- `TestNUF050ThinkingDeltaEmitsStructuredMessageUpdate`
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

Provider rate-limit errors retry up to five times before surfacing an error.
Interactive TUI renders `Rate limit` on the right side of the cwd/footer path
line while the status spinner shifts through an alert color gradient.

Tests:

- `TestAgentRetriesRateLimitBeforeFailing`
- `TestAgentStopsAfterFiveRateLimitRetries`
- `TestTUIRateLimitShowsFooterNotice`

## Messages And Events

### NUF-060 Message Types

Nu supports user, assistant, tool result, bash execution, custom, branch summary,
and compaction summary messages with text, image, thinking, and tool call
content blocks.
Interactive TUI message state keeps visible text, model thinking, and tool
execution blocks as ordered parts instead of one concatenated string.

Tests:

- `TestNUF060MessageJSONRoundTrip`
- `TestNUF060ImageContentRoundTrip`
- `TestTUIAppRendersStructuredMessageParts`

### NUF-061 Event Stream

Nu emits lifecycle, turn, message, tool execution, queue, compaction, retry, and
session events. JSON mode writes one event per line.
Message events distinguish visible text deltas from thinking deltas. Tool
events carry enough data for clients to render pending and completed tool
blocks. `rate_limit` events carry retry attempt metadata for status/footer
rendering.

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
truncated output to a temp file. Interactive `sudo` is rejected before process
start so password prompts cannot appear in the TUI input row; callers must use
`sudo -n`, `sudo -S`, or `sudo --non-interactive` explicitly.

Tests:

- `TestNUF073BashCapturesStdoutAndStderr`
- `TestNUF073BashTimeoutKillsProcess`
- `TestNUF073BashTruncatesAndPersistsFullOutput`
- `TestBashRejectsInteractiveSudo`
- `TestBashAllowsNonInteractiveSudoForms`

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

Sessions are append-only JSONL trees. The file header includes schema version,
session id, and cwd; each entry includes schema version, parent id, timestamp,
kind, and payload according to `protocols/session-jsonl.md`.

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
keybinding help, message history, bordered editor, cwd/context footer, status,
tool rendering, thinking rendering, images where supported, selectors, overlays,
and resize handling.

The message viewport autoscrolls to the bottom during normal streaming, but
PageUp/PageDown, End, and mouse wheel input can inspect older output without
moving the editor/footer away from the bottom anchor. A single reserved status
row stays directly above the editor and animates short state labels such as
`running`, `bubbling`, `running tool`, and `aborting`. The footer displays
estimated context usage against the selected context window until provider
token-usage events are available.

The first Go UI implementation must expose the same wiring boundaries as Pi:
terminal input becomes editor/command actions, agent events update component
state, and rendering produces deterministic frames that the terminal driver
writes in place. Raw terminal integration can stay narrow, but raw lifecycle,
diff rendering, input decoder, editor buffer, component subpackages, and
app-mode wiring are required before further interactive features build on them.

Tests:

- `TestNUF100RendererDoesNotOverflowWidth`
- `TestNUF100ResizeInvalidatesLayout`
- `TestTerminalDrawRepaintsWithANSI`
- `TestTUIAppRendersPiStyleComponentTree`
- `TestTUIAppAnchorsEditorAndFooterToBottom`
- `TestTUIAppKeepsStatusLineAboveEditor`
- `TestTUIHandleRawInputScrollsViewport`
- `TestEngineRendersComponentTreeAndDiffs`
- `TestEngineScrollsOverflowingViewport`
- `TestDecoderReadsPrintableUTF8AndEscapeSequence`
- `TestEditorHandlesUnicodeCursorAndPaste`
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

`/model` without arguments opens a Pi-style model selector. `/model <query>`
selects an exact model match when possible, otherwise opens the selector with
that query prefilled. Selector confirmation updates the footer/status and the
provider-backed agent used for later prompts.

Every built-in slash command has a local TUI handler. Commands that already
have enough local state mutate it directly (`/name`, `/new`, `/fork`,
`/compact`, `/reload`), file commands use current in-memory chat
(`/export`, `/import`, `/resume`, `/share`), auth/trust commands update the
same global files used by runtime wiring, and diagnostic commands render
Markdown output instead of falling through to a backend placeholder.

Tests:

- `TestBuiltinsCopiesPiCommandSet`
- `TestTUICommandMenuRendersAndCompletes`
- `TestTUISlashModelOpensSelectorAndSelectsModel`
- `TestTUISlashModelExactMatchSelectsWithoutMenu`
- `TestTUIAllBuiltinSlashCommandsHaveHandlers`
- `TestTUISlashSessionDoesNotCallAgent`
- `TestTUISlashQuitRequestsExit`

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

## Integrated Agent SDK Backend

### NUF-200 Curated Internal SDK Fork

Nu contains a curated fork sourced from `Ingenimax/agent-sdk-go` `v0.2.62`
under `internal/`. Every SDK feature and API behavior currently imported from
that pinned baseline remains available; this is a structural reorganization,
not a feature-reduction project. The final target is the exact balanced hierarchy
in NUA-011: `app/{auth,cli}`; `agent/{config,plans,guardrails,prompts}`;
`llm/{openai,anthropic,gemini,azureopenai,deepseek,ollama,vllm}`;
`tools/{coding,search,image,graphrag}`; `data/{embedding,weaviate/{graph,vector},sql,storage}`;
`task/{service,workflow,orchestration}`; `telemetry/{otel,langfuse}`;
`transport/{grpc/pb,http,a2a,ui}`; and
`tui/{core,editor,engine,input,message,terminal,components}`. Standalone packages
are exactly `agentui`, `config`, `contracts`, `memory`, `multitenancy`, `mcp`,
`model`, `rpc`, `session`, and `testkit`, plus `cmd/nu`. Superseded paths are
deleted. Old paths receive no compatibility wrappers, aliases, facade package,
or duplicate backend. No feature or API behavior is deleted. The pinned commit
and MIT license remain recorded in `THIRD_PARTY_NOTICES.md` and
`internal/AGENT_SDK_LICENSE`.

Tests:

- `go test ./internal/agent/... ./internal/contracts ./internal/llm/...`
- full imported SDK package tests;
- structural check that only the approved target package roots exist and the
  imported feature/test inventory is preserved;
- `TestSDKForkDoesNotImportNuOwnedPackages`

### NUF-201 SDK Agent Runtime

`internal/agent.Agent` is the sole model/tool agent runtime. Nu constructs it
with an SDK `LLM`, conversation memory, Nu coding tools, bounded tool iterations,
streaming enabled, and a diagnostics-safe logger.

Tests:

- upstream `internal/agent` tests;
- retained retry and telemetry tests under `internal/llm` and
  `internal/telemetry/...`;
- `TestPrintModeBuildsProviderFromCLI`;
- `TestSDKStreamMapsContentThinkingAndTools`.

### NUF-202 TUI Stream Adapter

`internal/agentui` owns only UI lifecycle state and event translation. It maps
SDK content, thinking, tool call/result, error, and completion events to the
existing TUI/RPC event shape. It must not call an LLM or execute a tool itself.

Tests:

- `TestSDKStreamMapsContentThinkingAndTools`;
- `TestAbortCancelsSDKRunner`;
- existing TUI structured-message tests.

### NUF-203 SDK Providers

Nu model selection constructs SDK OpenAI, Anthropic, Gemini, Claude-on-Bedrock,
and OpenAI-compatible clients from explicit Nu auth/settings. OpenAI-compatible
base URLs and Fireworks use the SDK OpenAI client. Provider clients never write
diagnostics to protocol stdout.

Tests:

- `TestPrintModeBuildsProviderFromCLI`;
- `TestPrintModeBuildsFireworksProviderFromGlobalModels`;
- `TestPrintModeBuildsProviderFromSettings`;
- all provider tests under their owning `internal/llm/*` packages.

### NUF-204 Nu Coding Tools On SDK

The seven Nu coding tools and `Builtins(cwd)` live together in
`internal/tools/coding`; imported SDK tools live in the cohesive
`internal/tools/{search,image,graphrag}` families. Root `internal/tools` owns
Registry, Calculator, shared helpers, and agent-as-tool orchestration without
re-exporting child packages. The Nu tools implement
`internal/contracts.Tool` and are
supplied to the SDK agent. Their filesystem/process behavior and cwd safety
remain unchanged; the old provider/tool-loop contracts and old `internal/tool`
import path do not remain.

Tests:

- `TestBuiltinsExposesEveryPhaseTwoTool`;
- `TestDefinitionsExposeBashSchema`;
- existing leaf-tool tests.

### NUF-210 SDK Conversation Memory

The active Nu agent uses the SDK bounded in-process conversation memory and Nu
supplies stable organization/conversation IDs. All other memory behavior remains
in `internal/memory`; embedding, vector store, datastore, storage, Redis, and
retrieval behavior lives in `internal/data/{embedding,weaviate/{graph,vector},sql,storage}`.
Embedding owns embedders and generic metadata evaluation; Weaviate owns distinct
GraphRAG `Store` and vector `Store` implementations in separate
`graph` and `vector` packages; SQL owns
`Postgres*` and `Supabase*` adapters; storage owns the `Storage` contract and
`Local*`/`GCS*` implementations. Root `internal/data` has no forwarding API.
GraphRAG tools consume the Weaviate option helpers from their owning package.
All remain available even when not wired into the Nu CLI.

Tests:

- all imported memory and `internal/data/...` tests.

### NUF-211 SDK MCP Client

Nu exposes configured MCP tools, prompts, and resources over its supported
client transports. All other MCP behavior imported from the pinned SDK,
including server, transport, sampling, and management surfaces, remains
available at the SDK layer. Nu headless modes keep SDK diagnostics off stdout.

Tests:

- all imported `internal/mcp` tests;
- `TestNUF170JSONModeStdoutIsOnlyJSONL`.

### NUF-212 Full Imported SDK Feature Retention

The imported SDK baseline includes built-in/image tools, guardrails, tracing,
agent-as-tool, sub-agents, execution plans, task services, orchestration,
workflows, A2A, gRPC, HTTP/microservice adapters, GraphRAG, embeddings, vector
stores, datastores, storage, structured output, remote config, multi-tenancy,
all imported provider clients, and every other source-backed feature currently
present from pinned `v0.2.62`. Structural work must preserve their API behavior
and owning test coverage. A Nu user-facing requirement is needed to expose a
feature through the product, not to keep it available in the SDK fork.

The required structure is exactly NUA-011. Root packages own shared types and
cross-subpackage orchestration only. Concrete behavior moves to the listed
cohesive subpackages; ordinary filenames such as `client.go` do not repeat their
package name. All reusable TUI components share `internal/tui/components` rather
than one package per component. Generated protobuf lives only in
`internal/transport/grpc/pb`. The concrete remote clients remain outside
`agent` to avoid cycles and are injected through transport-neutral contracts.
No feature is deleted and no old-path wrapper is permitted.

Tests:

- full imported SDK test set passes after package moves/merges;
- ownership check covers every final SDK package without limiting the feature
  set;
- structural check rejects every superseded package path and compatibility
  wrapper.

### NUF-213 Model Switching And Reset

TUI model switching rebuilds an SDK agent with the selected SDK LLM while
preserving the same SDK memory. `/new` clears the scoped SDK conversation.
Concurrent prompts remain rejected by the UI controller and abort cancels the
active SDK stream context.

Tests:

- `TestAbortCancelsSDKRunner`;
- RPC model/session command tests;
- TUI model selector tests.

### NUF-214 No Legacy Backend

The old Nu provider abstraction, provider-specific adapters, scripted-provider
testkit, and custom agent loop do not exist. `internal/agent` is the imported SDK
package; `internal/agentui` is the only Nu adapter between SDK streams and TUI/RPC.
No wrapper package preserves old Nu backend or removed upstream SDK paths,
including legacy `auth`, `cli`, `slash`, `interfaces`, and any package outside
the exact NUA-011 hierarchy. Approved child packages are real owners, not
wrappers for temporary flat-root APIs.

Tests:

- repository import check rejects `nu/internal/provider`;
- `go build ./cmd/nu`;
- `go test ./internal/agentui ./internal/app/... ./internal/rpc ./internal/tui/...`;
- package-root allowlist check for NUF-200.

### NUF-215 Runnable Agent SDK Examples

The repository includes concise `package main` examples for agent construction,
providers, tools, scoped memory, MCP configuration, tasks, and telemetry. Examples
import the approved root and cohesive subpackage APIs directly and do not
introduce wrappers or example-only frameworks. Credentials come only from
environment configuration. All examples except the research agent run without
provider requests or external services.

Tests:

- `go test ./examples/...`;
- local execution commands in `spec/examples.md`.
