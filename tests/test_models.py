"""Tests for schedule models."""

import json
from pathlib import Path

import pytest

from cnayp_bot.models import Schedule, ScheduleConfig


def test_schedule_model():
    """Test Schedule model parsing."""
    data = {
        "name": "Test Event",
        "description": "Test description",
        "voice_channel": "general",
        "notify_channel": "events",
        "days": ["monday", "friday"],
        "time": "18:00",
        "timezone": "America/Lima",
        "duration_minutes": 120,
    }
    schedule = Schedule.model_validate(data)

    assert schedule.name == "Test Event"
    assert schedule.days == ["monday", "friday"]
    assert schedule.duration_minutes == 120


def test_schedule_config_model():
    """Test ScheduleConfig model parsing."""
    data = {
        "digest_time": "8:00",
        "digest_channel": "events",
        "reminder_minutes": [120, 30, 15],
        "schedules": [
            {
                "name": "KCNA Session",
                "description": "Study session",
                "voice_channel": "K8s | KCNA",
                "notify_channel": "events",
                "days": ["monday"],
                "time": "18:00",
                "timezone": "America/Lima",
                "duration_minutes": 120,
            }
        ],
    }
    config = ScheduleConfig.model_validate(data)

    assert config.digest_time == "8:00"
    assert len(config.schedules) == 1
    assert config.reminder_minutes == [120, 30, 15]


def test_schedule_config_defaults():
    """Test ScheduleConfig default values."""
    config = ScheduleConfig.model_validate({})

    assert config.schedules == []
    assert config.digest_time == ""
    assert config.digest_channel == ""
    assert config.reminder_minutes == [60, 15]


def test_load_schedules_json(tmp_path: Path):
    """Test loading schedules from JSON file."""
    schedules_data = {
        "digest_time": "8:00",
        "digest_channel": "events",
        "reminder_minutes": [120, 30],
        "schedules": [
            {
                "name": "Test Event",
                "description": "Test",
                "voice_channel": "voice",
                "notify_channel": "text",
                "days": ["monday"],
                "time": "10:00",
                "timezone": "UTC",
                "duration_minutes": 60,
            }
        ],
    }

    config_file = tmp_path / "schedules.json"
    config_file.write_text(json.dumps(schedules_data))

    with config_file.open() as f:
        data = json.load(f)

    config = ScheduleConfig.model_validate(data)
    assert len(config.schedules) == 1
    assert config.schedules[0].name == "Test Event"
