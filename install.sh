#!/bin/bash

set -e

PLATFORM=${1:-linux}
TAG=${2:-latest}
INSTALL_PATH=${3:-/usr/local/bin}
IMAGE="patronc2/cli:$PLATFORM-$TAG"
BINARY_NAME="patron"
[ "$PLATFORM" = "windows" ] && BINARY_NAME="patron.exe"

echo "Pulling $IMAGE..."
docker pull $IMAGE

CID=$(docker create $IMAGE)
echo "Copying $BINARY_NAME to $INSTALL_PATH"
docker cp "$CID:/$BINARY_NAME" "$INSTALL_PATH/$BINARY_NAME"
docker rm "$CID" > /dev/null

chmod +x "$INSTALL_PATH/$BINARY_NAME"
echo "✅ Installed $BINARY_NAME to $INSTALL_PATH"
