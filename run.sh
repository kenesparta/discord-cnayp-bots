#!/bin/bash
set -e

IMAGE_NAME="cnayp-bot"

echo "Stopping and removing existing containers..."
CONTAINERS=$(docker ps -aq --filter ancestor="$IMAGE_NAME")
if [ -n "$CONTAINERS" ]; then
    docker stop $CONTAINERS
    docker rm $CONTAINERS
fi

# Also remove by name in case it exists
docker rm -f "$IMAGE_NAME" 2>/dev/null || true

echo "Removing old image..."
docker rmi "$IMAGE_NAME" 2>/dev/null || true

echo "Building fresh image..."
docker build --no-cache -t "$IMAGE_NAME" .

echo "Running container..."
docker run -d --name "$IMAGE_NAME" --env-file .env "$IMAGE_NAME"

echo "Container started. View logs with: docker logs -f $IMAGE_NAME"
