"""Webhook server for receiving Google Calendar push notifications."""

import asyncio
import logging
from collections.abc import Callable, Coroutine
from typing import Any

from aiohttp import web

from ..config import settings

logger = logging.getLogger(__name__)


class WebhookServer:
    """HTTP server to receive Google Calendar webhook notifications."""

    def __init__(self, on_calendar_change: Callable[[], Coroutine[Any, Any, None]]) -> None:
        """Initialize the webhook server.

        Args:
            on_calendar_change: Async callback to invoke when calendar changes.
        """
        self._on_calendar_change = on_calendar_change
        self._app = web.Application()
        self._runner: web.AppRunner | None = None
        self._setup_routes()

    def _setup_routes(self) -> None:
        """Set up HTTP routes."""
        self._app.router.add_post("/webhook", self._handle_webhook)
        self._app.router.add_get("/health", self._handle_health)

    async def _handle_webhook(self, request: web.Request) -> web.Response:
        """Handle incoming webhook from Google Calendar.

        Google sends notifications with special headers:
        - X-Goog-Channel-ID: The channel ID
        - X-Goog-Resource-State: "sync" for initial sync, "exists" for changes
        """
        channel_id = request.headers.get("X-Goog-Channel-ID", "")
        resource_state = request.headers.get("X-Goog-Resource-State", "")

        logger.info("Webhook received: channel=%s, state=%s", channel_id, resource_state)

        if resource_state == "sync":
            # Initial sync confirmation from Google
            logger.info("Watch channel sync confirmed")
            return web.Response(status=200)

        if resource_state == "exists":
            # Calendar has changes
            asyncio.create_task(self._on_calendar_change())
            return web.Response(status=200)

        return web.Response(status=200)

    async def _handle_health(self, request: web.Request) -> web.Response:
        """Health check endpoint."""
        return web.Response(text="OK", status=200)

    async def start(self) -> None:
        """Start the webhook server."""
        self._runner = web.AppRunner(self._app)
        await self._runner.setup()

        site = web.TCPSite(
            self._runner,
            host=settings.webhook_host,
            port=settings.webhook_port,
        )
        await site.start()

        logger.info(
            "Webhook server started on %s:%d",
            settings.webhook_host,
            settings.webhook_port,
        )

    async def stop(self) -> None:
        """Stop the webhook server."""
        if self._runner:
            await self._runner.cleanup()
            logger.info("Webhook server stopped")
