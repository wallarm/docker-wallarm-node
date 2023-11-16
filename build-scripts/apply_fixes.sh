#!/bin/bash

ARCH=${ARCH:-x86_64}
DOCKER_ARCH=${DOCKER_ARCH:-amd64}

if [ $ARCH == "aarch64" ]; then
    DOCKER_ARCH="arm64"
fi

BUILD_DIR="build/linux/${DOCKER_ARCH}"

sed -i \
    -e '/WALLARM_COMPONENT_NAME/d' \
    -e '/WALLARM_COMPONENT_VERSION/d' \
    -e '/SLAB_ALLOC_ARENA/d' \
    $BUILD_DIR/opt/wallarm/env.list

sed -i \
    -e '/FQDNLookup/s/no/true/' \
    -e '/DeleteSocket/aSocketGroup "wallarm"' \
    $BUILD_DIR/opt/wallarm/etc/collectd/wallarm-collectd.conf

sed -i -E \
    -e '/LUA/s/(.*)=(.*)/\1="\2"/' \
    $BUILD_DIR/opt/wallarm/env.list

cat conf/supervisord.conf.socat >> $BUILD_DIR/opt/wallarm/etc/supervisord.conf.filtering
