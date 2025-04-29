# Wallarm Docker Aio

Wallarm Docker Aio is an all-in-one Docker image that provides a complete Wallarm Node installation for API security. This image combines Nginx with Wallarm's security modules to protect your applications from various types of attacks.

## Overview

This Docker image includes:
- Alpine Linux as the base OS
- Nginx web server
- Wallarm Node security modules
- Various Nginx modules for enhanced functionality
- Gomplate for template processing

## Prerequisites

- Docker installed on your system
- Docker Buildx (for multi-architecture builds)
- Git (for development)

## Quick Start

To run the Wallarm Node container:

```bash
docker run -d \
  -p 80:80 \
  -p 443:443 \
  -e WALLARM_API_TOKEN='your-api-token' \
  -e WALLARM_API_HOST='api.wallarm.com' \
  wallarm/node:latest
```

## Building the Image

You can build the image using the provided Makefile:

```bash
# Build the image
make docker-image-build

# Build and push to registry
make docker-push

# Run smoke tests
make smoke-test
```

## Configuration

The container can be configured using environment variables:

- `WALLARM_API_TOKEN`: Your Wallarm API token
- `WALLARM_API_HOST`: Wallarm API host (default: api.wallarm.com)
- `NGINX_BACKEND`: Backend server address
- `WALLARM_MODE`: Operation mode (block, monitoring, off)

## Architecture Support

The image supports multiple architectures:
- x86_64 (amd64)
- ARM64 (aarch64)

## Version Information

Current versions used in the build:
- Alpine: 3.20
- Nginx: 1.26.3
- Wallarm AIO: 6.0.2
- Gomplate: 3.11.7

## Security

The container runs as a non-root user `wallarm` for enhanced security. The image includes:
- Regular security updates
- Minimal attack surface
- Proper capability management
- Secure default configurations

## Testing

The project includes smoke tests that can be run using:

```bash
make smoke-test
```

## Support

For support, please contact Wallarm Support Team at support@wallarm.com or visit our [documentation](https://docs.wallarm.com/installation/inline/compute-instances/docker/nginx-based).
