# Cross-Package Question Ledger

| ID | Status | Owner | Question | Safe default / closing evidence |
|---|---|---|---|---|
| `NUQ-001` | RESOLVED | architecture | SDK location | Curated SDK source is internalized directly as `internal/*`; no public module or compatibility tree |
| `NUQ-002` | PROVISIONAL | app/session | SDK memory to branchable session integration | Use bounded SDK conversation buffer; do not claim durable branch memory until adapter tests exist |
| `NUQ-003` | PROVISIONAL | model | OpenAI Responses versus SDK Chat path | Use imported SDK OpenAI implementation; expose exact API selection only after SDK supports and tests it |
| `NUQ-004` | PROVISIONAL | agentui | Retry/rate-limit event visibility | Preserve terminal errors; do not infer retry events from strings |
| `NUQ-005` | RESOLVED | backend | A2A version | Keep the A2A behavior imported from pinned SDK `v0.2.62`; reconsider its version only during a full baseline update |
| `NUQ-006` | PROVISIONAL | backend maintenance | Upgrade cadence | Stay on the pinned source commit until the full feature/API diff, structural patch ledger, generated output, and full tests are reviewed |
