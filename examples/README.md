# Agent SDK Examples

Run from the repository root:

| Example | Command |
| --- | --- |
| Research | `go run ./examples/research` |
| Providers | `go run ./examples/providers` |
| Tools | `go run ./examples/tools` |
| Memory | `go run ./examples/memory` |
| MCP | `go run ./examples/mcp` |
| Task | `go run ./examples/task` |
| Tracing | `go run ./examples/tracing` |

`research` calls OpenAI. Set `OPENAI_API_KEY`; optionally set `OPENAI_MODEL`,
`GOOGLE_API_KEY`, and `GOOGLE_SEARCH_ENGINE_ID` to enable web search.

`providers` only constructs clients and prints their names. Provider settings
come from `OPENAI_*`, `ANTHROPIC_*`, `GEMINI_API_KEY`, `AZURE_OPENAI_API_KEY`,
`AZURE_OPENAI_BASE_URL`, `AZURE_OPENAI_DEPLOYMENT`, `DEEPSEEK_API_KEY`,
`OLLAMA_*`, and `VLLM_*`. Gemini and Azure OpenAI are skipped when their
required settings are absent.

The other examples are local and require no credentials or network access.
They remain inside this module because they import `github.com/dm-vev/nu/internal/...` packages.
