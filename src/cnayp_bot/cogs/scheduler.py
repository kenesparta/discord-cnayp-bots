"""Scheduler cog for managing Discord events from Google Calendar."""

import logging
from datetime import datetime, timedelta
from zoneinfo import ZoneInfo

import discord
from discord.ext import commands, tasks

from ..config import settings
from ..services.calendar import CalendarEvent, CalendarService
from ..services.webhook import WebhookServer

logger = logging.getLogger(__name__)


class SchedulerCog(commands.Cog):
    """Manages Discord events and notifications from Google Calendar."""

    def __init__(self, bot: commands.Bot) -> None:
        self.bot = bot
        self.calendar = CalendarService()
        self.webhook_server: WebhookServer | None = None
        self.channel_cache: dict[str, int] = {}
        self.created_discord_events: set[str] = set()  # calendar_event_id
        self.sent_reminders: set[str] = set()  # "event_id:minutes"
        self.sent_start_notifications: set[str] = set()  # event_id
        self.known_events: dict[str, CalendarEvent] = {}  # event_id -> event

    async def cog_load(self) -> None:
        """Called when the cog is loaded."""
        if settings.webhook_enabled and settings.webhook_url:
            await self._start_webhook_mode()
        else:
            logger.info("Webhook disabled, using polling mode")

        self.scheduler_loop.start()
        self.reminder_loop.start()

    async def cog_unload(self) -> None:
        """Called when the cog is unloaded."""
        self.scheduler_loop.cancel()
        self.reminder_loop.cancel()

        if self.webhook_server:
            self.calendar.stop_watch()
            await self.webhook_server.stop()

    async def _start_webhook_mode(self) -> None:
        """Start webhook server and set up calendar watch."""
        logger.info("Starting webhook mode")

        # Start webhook server
        self.webhook_server = WebhookServer(on_calendar_change=self._on_calendar_change)
        await self.webhook_server.start()

        # Set up watch channel
        watch = self.calendar.setup_watch(settings.webhook_url)
        if watch:
            logger.info("Watch channel active until %s", watch.expiration)
        else:
            logger.warning("Failed to set up watch, falling back to polling")

    async def _on_calendar_change(self) -> None:
        """Handle calendar change notification from webhook."""
        logger.info("Calendar change detected via webhook")
        await self._process_calendar_changes()

    async def _process_calendar_changes(self) -> None:
        """Fetch and process calendar changes."""
        events = self.calendar.get_changes()

        for event in events:
            self.known_events[event.id] = event
            await self.check_and_create_discord_event(event)

    @tasks.loop(minutes=1)
    async def scheduler_loop(self) -> None:
        """Main scheduler loop for fetching events."""
        if settings.webhook_enabled and settings.webhook_url:
            # In webhook mode, only renew watch if needed
            await self._check_watch_renewal()
        else:
            # Polling mode: fetch events directly
            events = self.calendar.get_upcoming_events(hours_ahead=48)
            for event in events:
                self.known_events[event.id] = event
                await self.check_and_create_discord_event(event)

    @tasks.loop(minutes=1)
    async def reminder_loop(self) -> None:
        """Check for reminders and start notifications."""
        for event in self.known_events.values():
            await self.check_and_send_reminder(event)
            await self.check_and_send_start_notification(event)

    async def _check_watch_renewal(self) -> None:
        """Renew watch channel if it's about to expire."""
        watch = self.calendar.get_watch_channel()
        if not watch:
            # No watch active, try to set one up
            self.calendar.setup_watch(settings.webhook_url)
            return

        now = datetime.now(ZoneInfo("UTC"))
        time_until_expiry = watch.expiration - now

        # Renew if less than 1 hour until expiry
        if time_until_expiry < timedelta(hours=1):
            logger.info("Watch channel expiring soon, renewing")
            self.calendar.stop_watch()
            self.calendar.setup_watch(settings.webhook_url)

    @scheduler_loop.before_loop
    async def before_scheduler_loop(self) -> None:
        """Wait for bot to be ready before starting the loop."""
        await self.bot.wait_until_ready()

        # Initial fetch to populate known events
        events = self.calendar.get_upcoming_events(hours_ahead=48)
        for event in events:
            self.known_events[event.id] = event

        mode = "webhook" if settings.webhook_enabled else "polling"
        logger.info("Scheduler started in %s mode with %d events", mode, len(events))

    @reminder_loop.before_loop
    async def before_reminder_loop(self) -> None:
        """Wait for scheduler to start first."""
        await self.scheduler_loop.get_task()

    async def resolve_channel_id(self, channel_name: str) -> int | None:
        """Resolve a channel name to its ID, with caching."""
        if channel_name in self.channel_cache:
            return self.channel_cache[channel_name]

        guild = self.bot.get_guild(settings.discord_guild_id)
        if not guild:
            logger.error("Guild not found: %d", settings.discord_guild_id)
            return None

        for channel in guild.channels:
            self.channel_cache[channel.name] = channel.id

        return self.channel_cache.get(channel_name)

    async def check_and_create_discord_event(self, event: CalendarEvent) -> None:
        """Create a Discord scheduled event if not already created."""
        if event.id in self.created_discord_events:
            return

        now = datetime.now(ZoneInfo("UTC"))
        time_until_event = event.start_time - now

        if time_until_event > timedelta(hours=24) or time_until_event < timedelta(minutes=0):
            return

        voice_channel_id = await self.resolve_channel_id(settings.discord_voice_channel)
        if not voice_channel_id:
            logger.error("Failed to resolve voice channel: %s", settings.discord_voice_channel)
            return

        notify_channel_id = await self.resolve_channel_id(settings.discord_notify_channel)
        if not notify_channel_id:
            logger.error("Failed to resolve notify channel: %s", settings.discord_notify_channel)
            return

        guild = self.bot.get_guild(settings.discord_guild_id)
        if not guild:
            logger.error("Guild not found")
            return

        voice_channel = guild.get_channel(voice_channel_id)
        if not voice_channel:
            logger.error("Voice channel not found")
            return

        try:
            discord_event = await guild.create_scheduled_event(
                name=event.name,
                description=event.description or "Event from Google Calendar",
                start_time=event.start_time,
                end_time=event.end_time,
                channel=voice_channel,
                privacy_level=discord.PrivacyLevel.guild_only,
            )
            self.created_discord_events.add(event.id)
            logger.info("Created Discord event: %s (starts %s)", event.name, event.start_time)
        except discord.HTTPException as e:
            logger.error("Failed to create Discord event: %s", e)
            return

        notify_channel = self.bot.get_channel(notify_channel_id)
        if not notify_channel:
            return

        notification = (
            f"**New Event Alert!**\n\n"
            f"**{event.name}**\n"
            f"{event.description}\n\n"
            f"**When:** <t:{int(event.start_time.timestamp())}:F> (<t:{int(event.start_time.timestamp())}:R>)\n"
            f"**Timezone:** {event.timezone}\n"
            f"**Duration:** {event.duration_minutes} minutes\n"
            f"**Where:** <#{voice_channel_id}>\n\n"
            f"See you there!\n"
            f"https://discord.com/events/{settings.discord_guild_id}/{discord_event.id}"
        )

        await notify_channel.send(notification, allowed_mentions=discord.AllowedMentions(everyone=True))
        logger.info("Sent event notification for: %s", event.name)

    async def check_and_send_reminder(self, event: CalendarEvent) -> None:
        """Send reminder if we're at a reminder interval."""
        now = datetime.now(ZoneInfo("UTC"))
        minutes_until = int((event.start_time - now).total_seconds() / 60)

        for reminder_mins in settings.reminder_minutes:
            reminder_key = f"{event.id}:{reminder_mins}"

            if reminder_key in self.sent_reminders:
                continue

            if abs(minutes_until - reminder_mins) <= 1:
                await self.send_reminder(event, reminder_mins)
                self.sent_reminders.add(reminder_key)

    async def send_reminder(self, event: CalendarEvent, minutes_before: int) -> None:
        """Send a reminder for an upcoming event."""
        notify_channel_id = await self.resolve_channel_id(settings.discord_notify_channel)
        if not notify_channel_id:
            return

        channel = self.bot.get_channel(notify_channel_id)
        if not channel:
            return

        voice_channel_id = await self.resolve_channel_id(settings.discord_voice_channel)

        if minutes_before >= 60:
            hours = minutes_before // 60
            time_text = "1 hour" if hours == 1 else f"{hours} hours"
        else:
            time_text = f"{minutes_before} minutes"

        msg = (
            f"**Reminder:** {event.name} starts in {time_text}!\n\n"
            f"**Duration:** {event.duration_minutes} minutes\n"
            f"{event.description}\n\n"
            f"Join us in <#{voice_channel_id}>"
        )

        await channel.send(msg, allowed_mentions=discord.AllowedMentions(everyone=True))
        logger.info("Sent %s reminder for %s", time_text, event.name)

    async def check_and_send_start_notification(self, event: CalendarEvent) -> None:
        """Send notification when event is starting."""
        if event.id in self.sent_start_notifications:
            return

        now = datetime.now(ZoneInfo("UTC"))
        minutes_until = int((event.start_time - now).total_seconds() / 60)

        if abs(minutes_until) <= 1:
            await self.send_start_notification(event)
            self.sent_start_notifications.add(event.id)

    async def send_start_notification(self, event: CalendarEvent) -> None:
        """Send notification that an event is starting."""
        notify_channel_id = await self.resolve_channel_id(settings.discord_notify_channel)
        if not notify_channel_id:
            return

        channel = self.bot.get_channel(notify_channel_id)
        if not channel:
            return

        voice_channel_id = await self.resolve_channel_id(settings.discord_voice_channel)

        msg = (
            f"**{event.name} is starting now!**\n\n"
            f"{event.description}\n\n"
            f"**Duration:** {event.duration_minutes} minutes\n"
            f"**Timezone:** {event.timezone}\n\n"
            f"Join us in <#{voice_channel_id}>"
        )

        await channel.send(msg, allowed_mentions=discord.AllowedMentions(everyone=True))
        logger.info("Sent start notification for %s", event.name)


async def setup(bot: commands.Bot) -> None:
    """Set up the scheduler cog."""
    await bot.add_cog(SchedulerCog(bot))
