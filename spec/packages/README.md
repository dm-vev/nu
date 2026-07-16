# Package File Specs

This tree tracks active planned or implemented Go files only. Add a file spec
immediately before starting a file, not months before the package is ready.

Balanced hierarchy migration status: **IMPLEMENTED**. Existing file specs that
still name temporary flat-root files describe current source only; they do not
approve that path as the final owner. Move/update each affected file spec to its
approved path before moving the corresponding Go file.

Exception: the curated SDK fork under `internal/` is tracked by
`spec/backend.md`, `spec/sdk/README.md`, imported upstream tests, license, and
patch ledger. Do not create duplicate file specs for unchanged imported SDK
files. Nu-owned adapters and modifications still require file specs here; an SDK
file over the production line limit records its exception in `spec/backend.md`.

Each file spec must include:

- purpose;
- current status;
- TODO checklist;
- implementation commit and comments;
- code style;
- functions/types owned by the file;
- function logic;
- acceptance criteria;
- tests that prove the contract.

Implementation rule: do not add a Go file or exported function before its spec
exists here.

Readiness rule: before the first Go code in a subpackage, re-open every file
spec in that subpackage and confirm:

- file boundaries still match the protocol specs and phase goal;
- each function has concrete step-by-step logic, not acceptance criteria copied
  as implementation;
- tests named in the file prove the risky behavior first;
- any split/merge/rename is applied to `spec/packages/*` before Go files are
  created.

## Status Values

- `TODO`: not implemented.
- `IN_PROGRESS`: implementation branch/worktree has started.
- `IMPLEMENTED_UNCOMMITTED`: code and listed tests exist in the worktree, but no
  implementation commit can yet be recorded.
- `IMPLEMENTED`: code exists and tests listed in the file pass.
- `BLOCKED`: cannot proceed without a decision or external dependency.

When a file becomes `IMPLEMENTED`, update that file spec:

- set `Current: IMPLEMENTED`;
- replace `Implementation Commit: -` with the commit hash that introduced or
  completed the implementation;
- replace `Implementation Comments` with short implementation notes, tradeoffs,
  or follow-up risks;
- check off completed TODO items.

If implementation spans multiple commits, record the final commit that made the
listed tests pass and mention earlier commits in comments only if needed.

New SDK architecture documents do not require speculative file specs. Add the
smallest public/private file set immediately before each roadmap phase starts,
then move it through these statuses.

Common style for all Go files:

- keep package names lowercase and short;
- pass `context.Context` into blocking, streaming, process, network, and storage
  functions;
- return errors, do not panic for recoverable failures;
- wrap errors with operation context;
- avoid global mutable state;
- define interfaces in the consuming package;
- use table tests for parsers, matchers, and protocol conversion;
- add short in-function intent comments before non-trivial branches, protocol
  steps, locks, filesystem/process/network side effects, and deliberate
  simplifications;
- do not comment obvious assignments or restate the next line of code;
- keep `cmd/nu` thin.

## Package And File Boundaries

The exhaustive target is:

```text
app/{auth,cli}
agent/{config,plans,guardrails,prompts}
llm/{openai,anthropic,gemini,azureopenai,deepseek,ollama,vllm}
tools/{agent,calculator,registry,coding,search,image,graphrag}
memory/{conversation,history,redis,vector,factory}
mcp/{builder,client,config,fault,lazy,preset,prompt,registry,resource,retry,sampling,schema,tool,transport}
data/{embedding,weaviate/{graph,vector},sql,storage}
task/{service,workflow,orchestration}
telemetry/{otel,langfuse}
transport/{remote,grpc/{client,server,microservice,pb},http/server,a2a/{card,client,server,tool},ui/{server,trace}}
tui/{core,editor,engine,input,message,terminal,components}
standalone: agentui config contracts multitenancy model rpc session testkit
```

- A package is one domain or dependency boundary. Do not preserve an upstream
  package solely for layout compatibility.
- Root packages own shared types and cross-subpackage orchestration only.
- A subpackage must own a cohesive feature family. Do not create one for a
  helper that belongs with its caller.
- Inside a subpackage, use normal filenames such as `client.go`, `stream.go`, and
  `client_test.go`; do not repeat the package/provider name in the filename.
- Put every reusable TUI component in `internal/tui/components`. Do not create
  child packages per component.
- During migration, move behavior only into an approved owner and delete
  superseded paths after tests move. Changing the target package set requires a
  new architecture decision. Preserve behavior/tests and add no wrappers.
- A one-file package is valid only when it is itself a cohesive real boundary,
  never when it contains one extracted helper.
- A production file has one cohesive responsibility. Behavior-neutral splits
  within the same package are allowed and keep the package's existing tests.
- Split each non-generated production `.go` file over 300 lines or document a
  cohesion-based exception in its file spec. Generated and test files are
  excluded from this limit.
- Regenerate protobuf from source; never specify or perform hand edits to
  generated Go or descriptor bytes.
