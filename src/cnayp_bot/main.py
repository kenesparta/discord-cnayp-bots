import asyncio
import logging
import signal

from .bot import create_bot
from .config import settings

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
)
logging.getLogger("google_auth_httplib2").setLevel(logging.ERROR)
logger = logging.getLogger(__name__)


async def main() -> None:
    bot = create_bot()

    loop = asyncio.get_running_loop()
    stop_event = asyncio.Event()

    def signal_handler() -> None:
        logger.info("Received shutdown signal")
        stop_event.set()

    for sig in (signal.SIGINT, signal.SIGTERM):
        loop.add_signal_handler(sig, signal_handler)

    async def run_bot() -> None:
        try:
            await bot.start(settings.discord_bot_token)
        except asyncio.CancelledError:
            pass
        finally:
            if not bot.is_closed():
                await bot.close()

    bot_task = asyncio.create_task(run_bot())

    await stop_event.wait()

    logger.info("Shutting down bot...")
    bot_task.cancel()

    try:
        await bot_task
    except asyncio.CancelledError:
        pass

    logger.info("Bot shutdown complete")
