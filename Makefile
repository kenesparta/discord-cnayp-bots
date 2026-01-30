-include .env
export

.PHONY: install run test lint format clean docker-build docker-run

install:
	uv sync

install-dev:
	uv sync --all-extras

run:
	uv run python -m cnayp_bot

test:
	uv run pytest

test-cov:
	uv run pytest --cov=src/cnayp_bot --cov-report=term-missing

lint:
	uv run ruff check .

format:
	uv run ruff format .

check: lint
	uv run ruff format --check .

clean:
	rm -rf .venv __pycache__ .pytest_cache .ruff_cache
	find . -type d -name "__pycache__" -exec rm -rf {} + 2>/dev/null || true
	find . -type f -name "*.pyc" -delete 2>/dev/null || true

docker-build:
	docker build -t cnayp-bot .

docker-run: docker-build
	docker run -d --env-file .env cnayp-bot
