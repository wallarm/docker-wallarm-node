#!/bin/bash
set -e

ARCH=${ARCH:-x86_64}
DOCKER_ARCH=${DOCKER_ARCH:-amd64}

if [ "$ARCH" == "aarch64" ]; then
    DOCKER_ARCH="arm64"
fi

if [[ "$OSTYPE" == Darwin* ]]; then
  if ! command -v gnu-sed &> /dev/null
  then
    echo "gnu-sed could not be found. Run \"brew install gnu-sed\""
    echo "export PATH=\"/opt/homebrew/opt/gnu-sed/libexec/gnubin:\$PATH\""
    exit 1
  fi
fi

BUILD_DIR="build/linux/${DOCKER_ARCH}"

sed -i -E \
    -e '/WALLARM_COMPONENT_NAME/s/(.*)=(.*)/\1=wallarm-nginx-docker/' \
    -e "/WALLARM_COMPONENT_VERSION/s/(.*)=(.*)/\1=$AIO_VERSION/" \
    -e '/SLAB_ALLOC_ARENA/d' \
    "$BUILD_DIR/opt/wallarm/env.list"

sed -i \
    -e '/FQDNLookup/s/no/true/' \
    -e '/DeleteSocket/aSocketGroup "wallarm"' \
    "$BUILD_DIR/opt/wallarm/etc/collectd/wallarm-collectd.conf"

sed -i -E \
    -e '/LUA/s/(.*)=(.*)/\1="\2"/' \
    "$BUILD_DIR/opt/wallarm/env.list"

cat conf/supervisord.conf.socat >> "$BUILD_DIR/opt/wallarm/etc/supervisord.conf.filtering"

mkdir -p "$BUILD_DIR/usr/lib/nginx/modules" && \
    mv "$BUILD_DIR/opt/wallarm/modules/${WLRM_FOLDER}/"* "$BUILD_DIR/usr/lib/nginx/modules/" && \
    rm -rf "$BUILD_DIR/opt/wallarm/modules"
