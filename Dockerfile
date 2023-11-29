ARG ALPINE_VERSION

FROM alpine:${ALPINE_VERSION}

ARG ALPINE_VERSION
ARG NGINX_VERSION
ARG GOMPLATE_VERISON

MAINTAINER Wallarm Support Team <support@wallarm.com>

LABEL org.opencontainers.image.title="Docker official image for Wallarm Node. API security platform agent"
LABEL org.opencontainers.image.documentation="https://docs.wallarm.com/installation/inline/compute-instances/docker/nginx-based"
LABEL org.opencontainers.image.source="https://github.com/wallarm/docker-wallarm-node"
LABEL org.opencontainers.image.vendor="Wallarm"
LABEL org.opencontainers.image.version="${AIO_VERSION}"
LABEL org.opencontainers.image.revision="${CONTAINER_VERSION}"
LABEL com.wallarm.docker.versions.alpine-version="${ALPINE_VERSION}"
LABEL com.wallarm.docker.versions.nginx-version="${NGINX_VERSION}"

# core deps
RUN addgroup -S wallarm && \
    adduser -S -D -G wallarm -h /opt/wallarm wallarm && \
    apk update && \
    apk upgrade && \
    apk add curl bash socat logrotate libgcc "gomplate=~${GOMPLATE_VERISON}" && \
    curl -o /etc/apk/keys/nginx_signing.rsa.pub https://nginx.org/keys/nginx_signing.rsa.pub && \
    apk add -X "https://nginx.org/packages/mainline/alpine/v${ALPINE_VERSION}/main" "nginx=~${NGINX_VERSION}" "nginx-module-geoip=~${NGINX_VERSION}" "nginx-module-image-filter=~${NGINX_VERSION}" "nginx-module-perl=~${NGINX_VERSION}" "nginx-module-xslt=~${NGINX_VERSION}" && \
    nginx -v && \
    rm -r /var/cache/apk/*

# install wallarm
ARG WLRM_FOLDER
ARG TARGETPLATFORM
COPY --chown=wallarm:wallarm build/$TARGETPLATFORM/ /

# build-time compat check
COPY build-scripts/check_sig.sh /opt/wallarm/check_sig.sh
RUN /bin/sh -c '/opt/wallarm/check_sig.sh' && rm /opt/wallarm/check_sig.sh

# init script
COPY scripts/init /usr/local/bin/

# configs
RUN /bin/bash -c \
    'mkdir -p /etc/nginx/{modules-available,sites-available,sites-enabled} && \
    ln -sf /etc/nginx/modules/ /etc/nginx/modules-enabled && \
    ln -sf /etc/nginx/modules-available/mod-http-wallarm.conf /etc/nginx/modules-enabled/ && \
    rm /etc/nginx/conf.d/{default,stream}.conf || true && \
    touch /etc/environment && \
    chown -R wallarm:wallarm /run /etc/environment /etc/nginx /var/log/nginx /var/cache/nginx'
COPY conf/nginx /etc/nginx/
COPY conf/nginx_templates /opt/wallarm/

EXPOSE 80 443
USER wallarm

CMD ["/usr/local/bin/init"]
