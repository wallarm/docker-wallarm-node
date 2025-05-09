FROM alpine:latest

WORKDIR /app

ENV GO_VERSION=1.21.6 \
    ALLURECTL_VERSION=0.1.8 \
    VAULT_VERSION=1.15.5

RUN apk add --no-cache \
    curl \
    git \
    bash \
    ca-certificates \
    docker-cli \
    openssh \
    make \
    gcc \
    musl-dev

# Install Go
RUN curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" | tar -C /usr/local -xz && \
    ln -s /usr/local/go/bin/go /usr/bin/go && \
    ln -s /usr/local/go/bin/gofmt /usr/bin/gofmt

# Install allurectl
RUN wget https://github.com/allure-framework/allurectl/releases/latest/download/allurectl_linux_amd64 -O /usr/local/allurectl && \
    chmod +x /usr/local/allurectl && \
    ln -s /usr/local/allurectl /usr/bin/allurectl

# Install Vault CLI
RUN curl -fsSL "https://releases.hashicorp.com/vault/${VAULT_VERSION}/vault_${VAULT_VERSION}_linux_amd64.zip" -o vault.zip && \
    unzip vault.zip && \
    mv vault /usr/local/bin/ && \
    chmod +x /usr/local/bin/vault && \
    rm vault.zip

SHELL ["/bin/bash", "-c"]
CMD ["/bin/sh"]

