# Nu Agent Run Flow

```mermaid
sequenceDiagram
    participant U as User/TUI/RPC
    participant C as internal/agentui
    participant A as agent SDK
    participant L as internal/llm SDK
    participant T as SDK/Nu tools
    participant M as SDK memory

    U->>C: Prompt(ctx, text)
    C->>C: reject busy, own cancel
    C->>A: RunStream(ctx with Nu scope, text)
    A->>M: append/load conversation
    A->>L: GenerateWithToolsStream
    L-->>A: content/thinking/tool events
    A->>T: Execute tool calls
    T-->>A: results
    A-->>C: AgentStreamEvent
    C-->>U: Nu TUI/RPC Event
```

`internal/agentui` never calls `internal/llm` or tools. SDK Agent owns all loop
continuations. Abort cancels the context. Model switch rebuilds Agent with the
same memory and tools only while idle.
