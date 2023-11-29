ARG ALPINE_VERSION

FROM alpine:${ALPINE_VERSION}

ARG ALPINE_VERSION
ARG NGINX_VERSION

MAINTAINER Wallarm Support Team <support@wallarm.com>
LABEL NGINX_VERSION=${NGINX_VERSION}
LABEL AIO_VERSION=${AIO_VERSION}

# core deps
RUN addgroup -S wallarm && \
    adduser -S -D -G wallarm -h /opt/wallarm wallarm && \
    apk update && \
    apk upgrade && \
    apk add curl bash socat logrotate libgcc gomplate && \
    curl -o /etc/apk/keys/nginx_signing.rsa.pub https://nginx.org/keys/nginx_signing.rsa.pub && \
    apk add -X "https://nginx.org/packages/mainline/alpine/v${ALPINE_VERSION}/main" --no-cache "nginx=~${NGINX_VERSION}" && \
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
    'mkdir -p /etc/nginx/{modules-available,modules-enabled,sites-available,sites-enabled} && \
    ln -sf /etc/nginx/modules-available/mod-http-wallarm.conf /etc/nginx/modules-enabled/ && \
    rm /etc/nginx/conf.d/default.conf && \
    touch /etc/environment && \
    chown -R wallarm:wallarm /run /etc/environment /etc/nginx /var/log/nginx /var/cache/nginx'
COPY conf/nginx /etc/nginx/
COPY conf/nginx_templates /opt/wallarm/

EXPOSE 80 443
USER wallarm

CMD ["/usr/local/bin/init"]
