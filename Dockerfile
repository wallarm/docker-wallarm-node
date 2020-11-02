FROM debian:buster
MAINTAINER Wallarm Support Team <support@wallarm.com>

ENV DEBIAN_FRONTEND noninteractive

COPY conf/sources.list /etc/apt/sources.list.d/wallarm-node.list

RUN apt-get -y update && apt-get -y upgrade \
    && apt-get -y --no-install-recommends install \
        cron \
        logrotate \
        monitoring-plugins \
        supervisor \
        nginx \
        wallarm-node \
        libnginx-mod-http-wallarm \
        collectd-utils \
        curl \
        iptables \
        bsdmainutils \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/* \
    && chown -R wallarm:wallarm /var/lib/wallarm-tarantool \
    && sed -i -e 's|/var/log/wallarm/brute\.log|/var/log/wallarm/brute-detect\.log|' /etc/logrotate.d/wallarm-common \
    && sed -i -e 's|/usr/share/wallarm-common/syncnode|/usr/share/wallarm-common/syncnode -c /etc/wallarm-dist/node.yaml|' /etc/cron.d/wallarm-node-nginx \
    && rm -rf /etc/wallarm/triggers.d/ \
    && mkdir -p /etc/wallarm-dist/triggers.d

COPY conf/node.yaml /etc/wallarm-dist/
COPY scripts/trigger /etc/wallarm-dist/triggers.d/nginx
COPY scripts/init /usr/local/bin/
COPY conf/supervisord.conf /etc/supervisor/
COPY conf/logrotate.conf /etc/
COPY conf/default /etc/nginx/sites-enabled/
COPY conf/wallarm-status.conf /etc/nginx/conf.d/
COPY conf/collectd.conf /etc/collectd/
COPY conf/wallarm-acl.conf /etc/nginx/conf.d/
COPY conf/nginx-blacklistonly.conf /etc/nginx/

EXPOSE 80 443

CMD ["/usr/local/bin/init"]
