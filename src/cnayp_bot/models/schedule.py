"""Schedule configuration models."""

from pydantic import BaseModel, Field


class Schedule(BaseModel):
    """A scheduled event configuration."""

    name: str
    description: str
    voice_channel: str
    notify_channel: str
    days: list[str]
    time: str
    timezone: str
    duration_minutes: int


class ScheduleConfig(BaseModel):
    """Root configuration for schedules."""

    schedules: list[Schedule] = Field(default_factory=list)
    digest_time: str = ""
    digest_channel: str = ""
    reminder_minutes: list[int] = Field(default_factory=lambda: [60, 15])
