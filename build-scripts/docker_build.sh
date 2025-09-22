#!/bin/bash

set -x
set -e
set -a

ARCH=$(uname -m)

# Single-platform for local, multi-platform for CI
if [[ "${CI:-false}" == "true" ]]; then
  PLATFORMS=${PLATFORMS:-linux/amd64,linux/aarch64}
  ARCHS=${ARCHS:-x86_64 aarch64}
  BUILDX_ARG=${BUILDX_ARGS:---push}
else
  if [ "$ARCH" = "x86_64" ]; then
    PLATFORMS=linux/amd64
    ARCHS=x86_64
  elif [ "$ARCH" = "arm64" ]; then
    PLATFORMS=linux/arm64
    ARCHS=aarch64
  elif [ "$ARCH" = "amd64" ]; then
    PLATFORMS=linux/amd64
    ARCHS=x86_64
  else
    >&2 echo "Unsupported architecture $ARCH"
    exit 1
  fi
  BUILDX_ARGS=--load
fi

echo "AIO version: $AIO_VERSION"

for ARCH in $ARCHS
do
  if [ ! "$SKIP_AIO_DOWNLOAD" = "true" ]; then
    ARCH=$ARCH build-scripts/get_dependencies.sh
  fi
  ARCH=$ARCH build-scripts/apply_fixes.sh
done

# setup buildx
docker buildx rm multi-arch || true
docker buildx create \
  --name multi-arch \
  --platform linux/amd64,linux/arm64 \
  --driver docker-container \
	--use

# run docker build
docker buildx build \
  --platform $PLATFORMS -f Dockerfile \
  --build-arg CONTAINER_VERSION="$CONTAINER_VERSION" \
  --build-arg GOMPLATE_VERISON="$GOMPLATE_VERISON" \
  --build-arg ALPINE_VERSION="$ALPINE_VERSION" \
  --build-arg NGINX_VERSION="$NGINX_VERSION" \
  --build-arg AIO_VERSION="$AIO_VERSION" \
  --build-arg COMMIT_SHA="$COMMIT_SHA" \
  --provenance=true \
  --sbom=true \
  --no-cache \
  -t $IMAGE $BUILDX_ARGS .
