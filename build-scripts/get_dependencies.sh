#!/bin/bash

set -x
set -e
ARCH=${ARCH:-x86_64}
DOCKER_ARCH=${DOCKER_ARCH:-amd64}

if [ $ARCH == "aarch64" ]; then
    DOCKER_ARCH="arm64"
fi

AIO_FILE="wallarm-${AIO_VERSION}.${ARCH}-musl.tar.gz"
AIO_URL="https://storage.googleapis.com/meganode_storage/${AIO_VERSION%.*}/${AIO_FILE}"
BUILD_DIR="build/linux/${DOCKER_ARCH}"

echo "Downloading AiO archive ${AIO_FILE} ..."
curl --create-dirs -L -C - -o "$BUILD_DIR/wallarm-${ARCH}.tar.gz" "$AIO_URL"

echo "Extracting into ${BUILD_DIR}"
tar -xf "$BUILD_DIR/wallarm-${ARCH}.tar.gz" -C "$BUILD_DIR"
