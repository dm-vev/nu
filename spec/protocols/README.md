# Protocol Specs

These files define formats that other packages must obey. They are stricter than
package specs because changing them affects persisted sessions, headless clients,
extensions, and provider adapters.

- `provider-stream.md`: provider-neutral event stream consumed by the agent.
- `session-jsonl.md`: persisted session file format.
- `rpc-jsonl.md`: headless RPC command/response/event protocol.
- `extension-jsonl.md`: out-of-process extension host protocol.
- `tui-rendering.md`: terminal frame/input invariants.

Rule: update the protocol spec and golden tests before changing any persisted or
wire-visible field.
