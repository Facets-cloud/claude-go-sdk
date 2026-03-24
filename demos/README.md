# Claude Agent SDK Demos (Go)

Go ports of the official [claude-agent-sdk-demos](https://github.com/anthropics/claude-agent-sdk-demos) by Anthropic.

> **Note:** These are community-maintained Go ports of the official TypeScript demos, not official Anthropic products.

## Available Demos

### [Hello World](./hello-world)
Simple getting-started example with a PreToolUse hook that restricts file writes.

```bash
go run ./demos/hello-world
```

### [Hello World V2](./hello-world-v2)
V2 Session API examples: basic session, multi-turn conversation, one-shot prompt, and session resume.

```bash
go run ./demos/hello-world-v2 basic
go run ./demos/hello-world-v2 multi-turn
go run ./demos/hello-world-v2 one-shot
go run ./demos/hello-world-v2 resume
```

### [Resume Generator](./resume-generator)
Web-searches a person and generates a professional 1-page .docx resume.

```bash
go run ./demos/resume-generator "Jane Doe"
```

### [Simple Chat App](./simple-chatapp)
Terminal-based multi-turn chat using streaming input channels.

```bash
go run ./demos/simple-chatapp
```

## Prerequisites

- Go 1.22+
- [Claude Code CLI](https://docs.claude.com/en/docs/claude-code) installed and authenticated
- `ANTHROPIC_API_KEY` environment variable set

## Upstream demos not ported

| Demo | Reason |
|---|---|
| [email-agent](https://github.com/anthropics/claude-agent-sdk-demos/tree/main/email-agent) | Full-stack app with IMAP, WebSocket UI, database — too complex for a direct port |
| [excel-demo](https://github.com/anthropics/claude-agent-sdk-demos/tree/main/excel-demo) | Electron desktop app |
| [ask-user-question-previews](https://github.com/anthropics/claude-agent-sdk-demos/tree/main/ask-user-question-previews) | React + Vite frontend with HTML preview rendering |
| [research-agent](https://github.com/anthropics/claude-agent-sdk-demos/tree/main/research-agent) | Python-based (uses `uv`, not TypeScript SDK) |
