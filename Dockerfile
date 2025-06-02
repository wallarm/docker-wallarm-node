ARG ALPINE_VERSION

FROM alpine:${ALPINE_VERSION}

ARG ALPINE_VERSION
ARG AIO_VERSION
ARG COMMIT_SHA
ARG CONTAINER_VERSION
ARG GOMPLATE_VERISON
ARG NGINX_VERSION
ARG TARGETPLATFORM
ARG TARGETARCH

LABEL org.opencontainers.image.authors="Wallarm Support Team <support@wallarm.com>"
LABEL org.opencontainers.image.title="Docker official image for Wallarm Node. API security platform agent"
LABEL org.opencontainers.image.documentation="https://docs.wallarm.com/installation/inline/compute-instances/docker/nginx-based"
LABEL org.opencontainers.image.source="https://github.com/wallarm/docker-wallarm-node"
LABEL org.opencontainers.image.vendor="Wallarm"
LABEL org.opencontainers.image.version="${CONTAINER_VERSION}"
LABEL org.opencontainers.image.revision="${COMMIT_SHA}"
LABEL com.wallarm.nginx-docker.versions.alpine="${ALPINE_VERSION}"
LABEL com.wallarm.nginx-docker.versions.nginx="${NGINX_VERSION}"
LABEL com.wallarm.nginx-docker.versions.aio="${AIO_VERSION}"
LABEL com.wallarm.nginx-docker.versions.gomplate="${GOMPLATE_VERISON}"

# core deps
RUN addgroup -S wallarm && \
    adduser -S -D -G wallarm -h /opt/wallarm wallarm && \
    apk update && \
    apk upgrade && \
    apk add curl bash logrotate libgcc binutils upx \
        grep \
        nginx=~$NGINX_VERSION \
        nginx-mod-http-perl=~$NGINX_VERSION \
        nginx-mod-stream=~$NGINX_VERSION \
        nginx-mod-http-dav-ext=~$NGINX_VERSION \
        nginx-mod-http-echo=~$NGINX_VERSION \
        nginx-mod-http-geoip=~$NGINX_VERSION \
        nginx-mod-http-image-filter=~$NGINX_VERSION \
        nginx-mod-mail=~$NGINX_VERSION \
        nginx-mod-http-upstream-fair=~$NGINX_VERSION \
        nginx-mod-http-xslt-filter=~$NGINX_VERSION && \
    nginx -V && \
    rm -r /var/cache/apk/*

# Download gomplate
RUN curl -sL https://github.com/hairyhenderson/gomplate/releases/download/v${GOMPLATE_VERISON}/gomplate_linux-${TARGETARCH} \
    -o /usr/local/bin/gomplate && \
    chmod 755 /usr/local/bin/gomplate && \
    upx /usr/local/bin/gomplate && \
    gomplate -v

# Create symlinks to redirect nginx logs to stdout and stderr
RUN ln -sf /dev/stdout /var/log/nginx/access.log && \
    ln -sf /dev/stderr /var/log/nginx/error.log

# install wallarm
COPY --chown=wallarm:wallarm build/$TARGETPLATFORM/ /

# auto-select matching module
COPY build-scripts/pick_module.sh /opt/wallarm/pick_module.sh
RUN /bin/sh /opt/wallarm/pick_module.sh && rm /opt/wallarm/pick_module.sh

# build-time compat check
COPY build-scripts/check_sig.sh /opt/wallarm/check_sig.sh
RUN /bin/sh -c '/opt/wallarm/check_sig.sh' && rm /opt/wallarm/check_sig.sh

# init script
COPY scripts/init /usr/local/bin/

# configs
RUN /bin/bash -c \
    'mkdir -p /etc/nginx/conf.d && \
    touch /etc/environment && \
    rm /etc/nginx/conf.d/stream.conf && \
    chown -R wallarm:wallarm /run /etc/environment /etc/nginx /var/log/nginx /var/lib/nginx'
COPY conf/nginx /etc/nginx/
COPY conf/nginx_templates /opt/wallarm/

RUN apk add --no-cache libcap && \
    setcap    cap_net_bind_service=+ep /opt/wallarm/usr/bin/wstore && \
    setcap -v cap_net_bind_service=+ep /opt/wallarm/usr/bin/wstore && \
    setcap    cap_net_bind_service=+ep /usr/sbin/nginx && \
    setcap -v cap_net_bind_service=+ep /usr/sbin/nginx && \
    apk del libcap binutils curl upx

EXPOSE 80 443
USER wallarm

CMD ["/usr/local/bin/init"]
