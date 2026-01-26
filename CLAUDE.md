# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Discord bots for CNCF (Cloud Native Computing Foundation) community, written in Go. This project uses **only the standard library** - no external Discord libraries. Interacts with Discord's HTTP REST API and WebSocket Gateway directly.

## Build and Run Commands

```bash
go build ./...          # Build the project
go build -o bot ./cmd/bot  # Build the bot binary
go test ./...           # Run tests
go test -run TestName ./path/to/package  # Run a specific test
go fmt ./...            # Format code
go vet ./...            # Vet code for issues

# Run the bot
DISCORD_BOT_TOKEN=your_token ./bot
```

## Go Version

This project uses Go 1.25.5. Prefer stdlib features available in Go 1.25 over external packages when possible.

## Code Quality

- **NEVER** use deprecated functions or packages. Always use the current recommended APIs.

## Architecture

```
cmd/bot/main.go           # Entry point, signal handling, graceful shutdown
internal/
  config/config.go        # Environment-based configuration
  discord/
    client.go             # REST API client (send messages, manage channels)
    gateway.go            # WebSocket connection (receive events, heartbeat)
    types.go              # Discord data structures (User, Message, Channel, Guild)
  bot/bot.go              # Bot logic, event handlers, command routing
```

### Key Components

- **Gateway**: Manages WebSocket connection to Discord. Handles opcodes (HELLO, HEARTBEAT, IDENTIFY, DISPATCH), maintains session, and dispatches events to registered handlers.
- **Client**: HTTP client for Discord REST API. Used to send messages and interact with Discord resources.
- **Bot**: Orchestrates gateway and client. Register handlers with `gateway.On("EVENT_NAME", handler)`.

### Adding New Features

1. Add event handler in `bot.go` using `b.gateway.On("EVENT_NAME", handlerFunc)`
2. Add REST API methods in `client.go` if needed
3. Add new types in `types.go` as required

## CRISP Code Directives

All code in this repository must follow the CRISP principles (https://bitfieldconsulting.com/posts/crisp-code):

### C - Correct
Code must do what the programmer intended. Approach code skeptically, assuming bugs exist. Tests are essential but not sufficient—they can be flawed too.

### R - Readable
Readability is what remains after removing obstacles to understanding. Prioritize clear variable names, consistent naming conventions, and logical flow.

### I - Idiomatic
Follow Go conventions and community standards. Use conventional patterns (`err` for errors, `r`/`w` for request/response, `ctx` for context). Learn idioms through studying quality Go code.

### S - Simple
Simplicity requires thought and effort. Favor directness over unnecessary abstractions. Don't pursue DRY so rigidly that it adds complexity. Write naturally within Go's paradigms.

### P - Performant
Performance matters least among these principles. Optimize for programmer time over CPU time in most cases, but maintain awareness of memory and computational efficiency.

## Security Directives

- **NEVER** read, display, output, or share the contents of `.env` files or any file containing tokens, secrets, or credentials.
- **NEVER** include real tokens, API keys, or secrets in code examples, commits, or responses.
- Treat `DISCORD_BOT_TOKEN` and all environment variables containing sensitive data as confidential.
