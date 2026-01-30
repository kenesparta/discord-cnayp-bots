# CNAYP Discord Bot

Discord bots for the CNAYP community with Google Calendar integration.

## Features

- Fetches events from Google Calendar
- Scheduled Discord event creation (24 hours in advance)
- Event reminders at configurable intervals (default: 60 and 15 minutes before)
- Event start notifications

## Setup

### 1. Install dependencies

```bash
uv sync
```

### 2. Set up Google Cloud

1. Go to [Google Cloud Console](https://console.cloud.google.com)
2. Create a project or select an existing one
3. Enable the **Google Calendar API**:
   - Go to **APIs & Services** → **Library**
   - Search for "Google Calendar API" and enable it

### 3. Configure Authentication

Choose one of the following methods:

#### Option A: Application Default Credentials (Recommended for GCP)

Best for running on GCP (Cloud Run, GKE, Compute Engine):

1. Create a Service Account:
   - Go to **IAM & Admin** → **Service Accounts**
   - Create a new service account
   - No need to download a key file

2. Assign the service account to your workload:
   - **Cloud Run**: Set the service account in the service settings
   - **GKE**: Use Workload Identity
   - **Compute Engine**: Set the service account on the VM

3. For local development:
   ```bash
   gcloud auth application-default login
   ```

#### Option B: Service Account JSON Key

Best for running outside GCP (AWS, local, etc.):

1. Create a Service Account:
   - Go to **APIs & Services** → **Credentials**
   - Click **Create Credentials** → **Service Account**
   - Click on the service account → **Keys** → **Add Key** → **Create new key** → **JSON**
   - Save it as `config/service-account.json`

2. Set the environment variable:
   ```bash
   GOOGLE_SERVICE_ACCOUNT_FILE=config/service-account.json
   ```

### 4. Get your Calendar ID

1. Go to Google Calendar
2. Click the three dots next to your calendar → **Settings and sharing**
3. Copy your **Calendar ID** from "Integrate calendar" section
   (looks like `abc123@group.calendar.google.com`)

> **Note:** If your calendar is **private**, you also need to share it with the service account email (under "Share with specific people" → "See all event details"). Public calendars don't require this.

### 5. Configure environment variables

```bash
cp .env.example .env
```

Edit `.env` with your credentials.

### 6. Run the bot

```bash
uv run python -m cnayp_bot
```

## Development

Run tests:
```bash
uv run pytest
```

Format code:
```bash
uv run ruff format .
```

Lint code:
```bash
uv run ruff check .
```

## Commands

- `!ping` - Check if the bot is responsive

## Configuration

Environment variables:

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DISCORD_BOT_TOKEN` | Yes | - | Your Discord bot token |
| `DISCORD_GUILD_ID` | Yes | - | Your Discord server/guild ID |
| `GOOGLE_CALENDAR_ID` | Yes | - | Your Google Calendar ID |
| `GOOGLE_SERVICE_ACCOUNT_FILE` | No | - | Path to service account JSON. If not set, uses ADC |
| `DISCORD_NOTIFY_CHANNEL` | No | `events` | Channel for notifications |
| `DISCORD_VOICE_CHANNEL` | No | `general` | Voice channel for events |
| `REMINDER_MINUTES` | No | `[60, 15]` | Minutes before event to send reminders |
