"""Configuration using Pydantic Settings."""

from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    """Bot configuration from environment variables."""

    model_config = SettingsConfigDict(env_file=".env", env_file_encoding="utf-8", extra="ignore")

    discord_bot_token: str
    discord_guild_id: int
    discord_notify_channel: str = "events"
    discord_voice_channel: str = "K8s | KCNA"

    google_calendar_id: str
    google_service_account_file: str | None = None

    # Webhook settings for real-time calendar notifications
    webhook_enabled: bool = False
    webhook_host: str = "0.0.0.0"
    webhook_port: int = 8080
    webhook_url: str | None = None  # Public URL (e.g., https://your-domain.com/webhook)

    reminder_minutes: list[int] = [90, 60, 15, 5]


settings = Settings()
