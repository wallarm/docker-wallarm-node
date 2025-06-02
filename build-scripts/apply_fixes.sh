#!/bin/bash
set -e

ARCH=${ARCH:-x86_64}
DOCKER_ARCH=${DOCKER_ARCH:-amd64}

if [ "$ARCH" == "aarch64" ]; then
    DOCKER_ARCH="arm64"
fi

BUILD_DIR="build/linux/${DOCKER_ARCH}"

sed -i -E \
    -e '/WALLARM_COMPONENT_NAME/s/(.*)=(.*)/\1=wallarm-nginx-docker/' \
    -e "/WALLARM_COMPONENT_VERSION/s/(.*)=(.*)/\1=$AIO_VERSION/" \
    -e '/SLAB_ALLOC_ARENA/d' \
    "$BUILD_DIR/opt/wallarm/env.list"

