"""Google Calendar service for fetching events."""

import logging
import uuid
from dataclasses import dataclass
from datetime import datetime, timedelta
from pathlib import Path
from zoneinfo import ZoneInfo

import google.auth
from google.oauth2 import service_account
from googleapiclient.discovery import build

from ..config import settings

logger = logging.getLogger(__name__)

SCOPES = ["https://www.googleapis.com/auth/calendar.readonly"]


def _get_credentials():
    """Get Google credentials using service account file or ADC.

    Priority:
    1. Service account JSON file (if GOOGLE_SERVICE_ACCOUNT_FILE is set and exists)
    2. Application Default Credentials (ADC)
    """
    if settings.google_service_account_file:
        creds_path = Path(settings.google_service_account_file)
        if creds_path.exists():
            logger.info("Using service account file: %s", creds_path)
            return service_account.Credentials.from_service_account_file(
                str(creds_path), scopes=SCOPES
            )
        logger.warning("Service account file not found: %s, falling back to ADC", creds_path)

    logger.info("Using Application Default Credentials (ADC)")
    credentials, project = google.auth.default(scopes=SCOPES)
    return credentials


@dataclass
class CalendarEvent:
    """Represents a calendar event."""

    id: str
    name: str
    description: str
    start_time: datetime
    end_time: datetime
    timezone: str

    @property
    def duration_minutes(self) -> int:
        """Calculate event duration in minutes."""
        delta = self.end_time - self.start_time
        return int(delta.total_seconds() / 60)


@dataclass
class WatchChannel:
    """Represents a Google Calendar watch channel."""

    id: str
    resource_id: str
    expiration: datetime


class CalendarService:
    """Service for interacting with Google Calendar API."""

    def __init__(self) -> None:
        self._service = None
        self._sync_token: str | None = None
        self._watch_channel: WatchChannel | None = None

    def _get_service(self):
        """Get or create the Google Calendar service."""
        if self._service is None:
            credentials = _get_credentials()
            self._service = build("calendar", "v3", credentials=credentials)

        return self._service

    def get_upcoming_events(self, hours_ahead: int = 24) -> list[CalendarEvent]:
        """Fetch upcoming events from the calendar.

        Args:
            hours_ahead: How many hours ahead to look for events.

        Returns:
            List of CalendarEvent objects.
        """
        service = self._get_service()

        now = datetime.now(ZoneInfo("UTC"))
        time_min = now.isoformat()
        time_max = (now + timedelta(hours=hours_ahead)).isoformat()

        try:
            events_result = (
                service.events()
                .list(
                    calendarId=settings.google_calendar_id,
                    timeMin=time_min,
                    timeMax=time_max,
                    singleEvents=True,
                    orderBy="startTime",
                )
                .execute()
            )
        except Exception as e:
            logger.error("Failed to fetch calendar events: %s", e)
            return []

        events = events_result.get("items", [])
        return [self._parse_event(event) for event in events]

    def setup_watch(self, webhook_url: str) -> WatchChannel | None:
        """Set up a watch channel for calendar push notifications.

        Args:
            webhook_url: The HTTPS URL to receive notifications.

        Returns:
            WatchChannel object if successful, None otherwise.
        """
        service = self._get_service()
        channel_id = str(uuid.uuid4())

        try:
            response = (
                service.events()
                .watch(
                    calendarId=settings.google_calendar_id,
                    body={
                        "id": channel_id,
                        "type": "web_hook",
                        "address": webhook_url,
                    },
                )
                .execute()
            )

            expiration_ms = int(response["expiration"])
            expiration = datetime.fromtimestamp(expiration_ms / 1000, tz=ZoneInfo("UTC"))

            self._watch_channel = WatchChannel(
                id=response["id"],
                resource_id=response["resourceId"],
                expiration=expiration,
            )

            logger.info(
                "Watch channel created: %s (expires %s)",
                self._watch_channel.id,
                expiration,
            )
            return self._watch_channel

        except Exception as e:
            logger.error("Failed to set up watch channel: %s", e)
            return None

    def stop_watch(self) -> bool:
        """Stop the current watch channel."""
        if not self._watch_channel:
            return True

        service = self._get_service()

        try:
            service.channels().stop(
                body={
                    "id": self._watch_channel.id,
                    "resourceId": self._watch_channel.resource_id,
                }
            ).execute()

            logger.info("Watch channel stopped: %s", self._watch_channel.id)
            self._watch_channel = None
            return True

        except Exception as e:
            logger.error("Failed to stop watch channel: %s", e)
            return False

    def get_watch_channel(self) -> WatchChannel | None:
        """Get the current watch channel."""
        return self._watch_channel

    def get_changes(self) -> list[CalendarEvent]:
        """Get calendar changes since last sync.

        Uses sync tokens to efficiently fetch only changed events.

        Returns:
            List of new/updated CalendarEvent objects.
        """
        service = self._get_service()
        events = []

        try:
            request_params = {
                "calendarId": settings.google_calendar_id,
                "singleEvents": True,
            }

            if self._sync_token:
                request_params["syncToken"] = self._sync_token
            else:
                # Initial sync: get events from now onwards
                now = datetime.now(ZoneInfo("UTC"))
                request_params["timeMin"] = now.isoformat()

            page_token = None
            while True:
                if page_token:
                    request_params["pageToken"] = page_token

                response = service.events().list(**request_params).execute()

                for item in response.get("items", []):
                    # Skip cancelled events
                    if item.get("status") == "cancelled":
                        continue
                    events.append(self._parse_event(item))

                page_token = response.get("nextPageToken")
                if not page_token:
                    self._sync_token = response.get("nextSyncToken")
                    break

            logger.info("Fetched %d changed events", len(events))
            return events

        except Exception as e:
            # Sync token might be invalid, reset and do full sync
            if "Sync token" in str(e) or "410" in str(e):
                logger.warning("Sync token invalid, resetting")
                self._sync_token = None
                return self.get_changes()

            logger.error("Failed to fetch changes: %s", e)
            return []

    def _parse_event(self, event: dict) -> CalendarEvent:
        """Parse a Google Calendar event into a CalendarEvent object."""
        start = event["start"]
        end = event["end"]

        if "dateTime" in start:
            start_time = datetime.fromisoformat(start["dateTime"])
            end_time = datetime.fromisoformat(end["dateTime"])
            timezone = start.get("timeZone", "UTC")
        else:
            start_date = datetime.fromisoformat(start["date"])
            end_date = datetime.fromisoformat(end["date"])
            start_time = start_date.replace(tzinfo=ZoneInfo("UTC"))
            end_time = end_date.replace(tzinfo=ZoneInfo("UTC"))
            timezone = "UTC"

        return CalendarEvent(
            id=event["id"],
            name=event.get("summary", "Untitled Event"),
            description=event.get("description", ""),
            start_time=start_time,
            end_time=end_time,
            timezone=timezone,
        )
