# Use the official Debian 12 image from the Docker Hub
FROM debian:12

# WALLARM_NODE_MODE: either filtering, postanalytics or empty (empty = filtering + postanalytics)
ARG WALLARM_NODE_MODE
ENV WALLARM_NODE_MODE=$WALLARM_NODE_MODE

# Set environment variables to non-interactive to avoid prompts
ENV DEBIAN_FRONTEND=noninteractive

# Install required packages for adding repositories and fetching keys
RUN apt-get update && apt-get install -y \
    gnupg2 \
    ca-certificates \
    lsb-release \
    wget \
    --no-install-recommends

# Conditionally add the NGINX signing key, repository, and install NGINX if WALLARM_NODE_MODE is not postanalytics
RUN if [ "$WALLARM_NODE_MODE" != "postanalytics" ]; then \
        wget https://nginx.org/keys/nginx_signing.key && apt-key add nginx_signing.key && \
        printf "deb http://nginx.org/packages/debian/ $(lsb_release -cs) nginx\n" | tee /etc/apt/sources.list.d/nginx.list && \
        apt-get update && apt-get install -y \
        nginx \
        --no-install-recommends; \
    fi

# Remove unnecessary files and clean up the package list
RUN apt-get clean && rm -rf /var/lib/apt/lists/*

# Expose port 80
EXPOSE 80

# Install AiO in specified mode, node registration skipped with the --install-only argument.
COPY aio.sh /aio.sh
RUN /aio.sh -- -b --install-only ${WALLARM_NODE_MODE} && rm /aio.sh

# Copy entry point script.
COPY entrypoint.sh /entrypoint.sh

# Command to run when starting the container
ENTRYPOINT ["/entrypoint.sh"]
