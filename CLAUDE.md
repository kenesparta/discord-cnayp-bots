# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Discord bots for CNAYP community, written in Python. Uses **discord.py** for Discord interaction, **Google Calendar API** for event scheduling, and **Pydantic** for configuration validation.

## Build and Run Commands

```bash
# Install dependencies (using uv)
uv sync

# Run the bot
uv run python -m cnayp_bot

# Run tests
uv run pytest

# Run a specific test
uv run pytest tests/test_models.py::test_schedule_model

# Format code
uv run ruff format .

# Lint code
uv run ruff check .
```

## Python Version

This project uses Python 3.12+. Use modern Python features (type hints, pattern matching, etc.).

## Code Quality

- **NEVER** use deprecated functions or packages. Always use the current recommended APIs.
- Use type hints for all function signatures.
- Follow PEP 8 style guidelines.

## Architecture

```
src/cnayp_bot/
  __init__.py           # Package init
  __main__.py           # Entry: python -m cnayp_bot
  main.py               # Bootstrap, signal handling
  config.py             # Pydantic Settings for env vars
  bot.py                # Bot class with commands
  cogs/
    __init__.py
    scheduler.py        # Scheduler with tasks.loop(), Google Calendar integration
  services/
    __init__.py
    calendar.py         # Google Calendar API service
  models/
    __init__.py
    schedule.py         # Pydantic models
```

### Key Components

- **Bot**: Main `CNAYPBot` class extending `commands.Bot`. Handles Discord events and commands.
- **Config**: Uses Pydantic Settings to load and validate environment variables.
- **CalendarService**: Fetches events from Google Calendar API using service account credentials.
- **Scheduler Cog**: Manages scheduled events using `tasks.loop()`. Handles:
  - Fetching events from Google Calendar (every minute)
  - Event start notifications
  - Reminders at configured intervals (default: 60, 15 minutes)
  - Discord scheduled event creation (24h in advance)

### Adding New Features

1. For new commands: Add methods with `@commands.command()` decorator in `bot.py`
2. For new scheduled tasks: Add to `scheduler.py` cog
3. For new config: Add fields to `config.py` Settings class
4. For new data models: Add to `models/` directory

## CRISP Code Directives

All code in this repository must follow the CRISP principles (https://bitfieldconsulting.com/posts/crisp-code):

### C - Correct
Code must do what the programmer intended. Approach code skeptically, assuming bugs exist. Tests are essential but not sufficient—they can be flawed too.

### R - Readable
Readability is what remains after removing obstacles to understanding. Prioritize clear variable names, consistent naming conventions, and logical flow.

### I - Idiomatic
Follow Python conventions and community standards. Use conventional patterns (snake_case, type hints, async/await for I/O). Learn idioms through studying quality Python code.

### S - Simple
Simplicity requires thought and effort. Favor directness over unnecessary abstractions. Don't pursue DRY so rigidly that it adds complexity.

### P - Performant
Performance matters least among these principles. Optimize for programmer time over CPU time in most cases, but maintain awareness of async operations and blocking code.

## Security Directives

- **NEVER** read, display, output, or share the contents of `.env` files or any file containing tokens, secrets, or credentials.
- **NEVER** include real tokens, API keys, or secrets in code examples, commits, or responses.
- Treat `DISCORD_BOT_TOKEN`, Google service account credentials, and all environment variables containing sensitive data as confidential.
