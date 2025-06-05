FROM alpine:latest

WORKDIR /app

ENV GO_VERSION=1.23.6 \
    VAULT_VERSION=1.15.5 \
    ALLURECTL_VERSION=2.17.0 \
    GOPATH=/go \
    PATH=/usr/local/go/bin:/go/bin:$PATH \
    CGO_ENABLED=0

# Install Go and other dependencies
RUN apk add --no-cache \
        curl git bash ca-certificates docker-cli openssh make gcc musl-dev unzip \
 && curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" \
      | tar -C /usr/local -xz \
 && ln -s /usr/local/go/bin/go /usr/bin/go \
 && ln -s /usr/local/go/bin/gofmt /usr/bin/gofmt

COPY . .
RUN go mod tidy


# Install utils
RUN wget -q https://github.com/allure-framework/allurectl/releases/download/${ALLURECTL_VERSION}/allurectl_linux_amd64 \
        -O /usr/local/allurectl \
 && chmod +x /usr/local/allurectl \
 && ln -s /usr/local/allurectl /usr/bin/allurectl \
 && curl -fsSL "https://releases.hashicorp.com/vault/${VAULT_VERSION}/vault_${VAULT_VERSION}_linux_amd64.zip" \
        -o vault.zip \
 && unzip -qq vault.zip \
 && mv vault /usr/local/bin/ \
 && chmod +x /usr/local/bin/vault \
 && rm vault.zip

SHELL ["/bin/bash", "-c"]
CMD ["/bin/sh"]