"""CNAYP Discord Bot."""

import logging

import discord
from discord.ext import commands

from .config import settings
from .services.calendar import CalendarService

logger = logging.getLogger(__name__)


class CNAYPBot(commands.Bot):
    """Main bot class for CNAYP Discord."""

    def __init__(self) -> None:
        intents = discord.Intents.default()
        intents.message_content = True
        intents.guilds = True

        super().__init__(command_prefix="!", intents=intents)
        self.calendar = CalendarService()

    async def setup_hook(self) -> None:
        """Called when the bot is starting up."""
        await self.load_extension("cnayp_bot.cogs.scheduler")
        logger.info("Loaded scheduler cog")

    async def on_ready(self) -> None:
        """Called when the bot is ready."""
        logger.info("Bot is ready! Logged in as %s", self.user)
        logger.info("Connected to guild: %d", settings.discord_guild_id)


def create_bot() -> CNAYPBot:
    """Create and configure the bot instance."""
    bot = CNAYPBot()

    @bot.command()
    async def ping(ctx: commands.Context) -> None:
        """Respond with pong."""
        await ctx.send("Pong!")

    @bot.command(name="events")
    async def list_events(ctx: commands.Context, days: int = 7) -> None:
        """List upcoming events from Google Calendar.

        Usage: !events [days]
        Example: !events 14 (shows events for next 14 days)
        """
        hours = days * 24
        events = bot.calendar.get_upcoming_events(hours_ahead=hours)

        if not events:
            await ctx.send(f"No events scheduled in the next {days} days.")
            return

        embed = discord.Embed(
            title=f"Upcoming Events ({days} days)",
            color=discord.Color.blue(),
        )

        for event in events[:10]:  # Limit to 10 events
            time_str = f"<t:{int(event.start_time.timestamp())}:F>"
            relative_str = f"<t:{int(event.start_time.timestamp())}:R>"
            embed.add_field(
                name=event.name,
                value=f"{time_str}\n{relative_str}\nDuration: {event.duration_minutes} min",
                inline=False,
            )

        if len(events) > 10:
            embed.set_footer(text=f"Showing 10 of {len(events)} events")

        await ctx.send(embed=embed)

    return bot
