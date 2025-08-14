# BUILD TEST BINARY
FROM golang:1.23-alpine AS builder

ENV CGO_ENABLED=0
WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go test -c -o /app/register.test ./cmd/... -tags functional

# BUILD FINAL IMAGE
FROM alpine:latest

ENV VAULT_VERSION=1.15.5
ENV ALLURECTL_VERSION=2.17.0

WORKDIR /app

RUN apk add --no-cache \
    ca-certificates \
    git \
    bash \
    unzip \
    curl \
    docker-cli \
    make

RUN wget -q "https://github.com/allure-framework/allurectl/releases/download/${ALLURECTL_VERSION}/allurectl_linux_amd64" -O /usr/local/bin/allurectl && \
    chmod +x /usr/local/bin/allurectl

RUN curl -fSsL "https://releases.hashicorp.com/vault/${VAULT_VERSION}/vault_${VAULT_VERSION}_linux_amd64.zip" -o vault.zip && \
    unzip -qq vault.zip && \
    mv vault /usr/local/bin/ && \
    chmod +x /usr/local/bin/vault && \
    rm vault.zip

COPY --from=builder /app/register.test .
COPY Makefile /app/Makefile

RUN chmod +x ./register.test

SHELL ["/bin/bash", "-c"]
CMD ["/bin/sh"]
