FROM debian:buster
MAINTAINER Wallarm Support Team <support@wallarm.com>

ENV DEBIAN_FRONTEND noninteractive

COPY conf/sources.list /etc/apt/sources.list.d/wallarm-node.list

RUN printf -- "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" | base64 -d > /usr/share/keyrings/wallarm.gpg \
    && printf -- "mQENBE5OMmIBCAD+FPYKGriGGf7NqwKfWC83cBV01gabgVWQmZbMcFzeW+hMsgxHW6iimD0RsfZ9oEbfJCPG0CRSZ7ppq5pKamYs2+EJ8Q2ysOFHHwpGrA2C8zyNAs4IQxnZZIbETgcSwFtDun0XiqPwPZgyuXVm9PAbLZRbfBzm8wR/3SWygqZBBLdQk5TEfDR+Eny/M1RVR4xClECONF9UBB2ejFdI1LD45APbP2hsN/piFByU1t7yK2gpFyRt97WzGHn9MV5/TL7AmRPM4pcr3JacmtCnxXeCZ8nLqedoSuHFuhwyDnlAbu8I16O5XRrfzhrHRJFM1JnIiGmzZi6zBvH0ItfyX6ttABEBAAG0KW5naW54IHNpZ25pbmcga2V5IDxzaWduaW5nLWtleUBuZ2lueC5jb20+iQE+BBMBAgAoAhsDBgsJCAcDAgYVCAIJCgsEFgIDAQIeAQIXgAUCV2K1+AUJGB4fQQAKCRCr9b2Ce9m/YloaB/9XGrolkocm7l/tsVjaBQCteXKuwsm4XhCuAQ6YAwA1L1UheGOG/aa2xJvrXE8X32tgcTjrKoYoXWcdxaFjlXGTt6jV85qRguUzvMOxxSEM2Dn115etN9piPl0Zz+4rkx8+2vJGF+eMlruPXg/zd88NvyLq5gGHEsFRBMVufYmHtNfcp4okC1klWiRIRSdp4QY1wdrN1O+/oCTl8Bzy6hcHjLIq3aoumcLxMjtBoclc/5OTioLDwSDfVx7rWyfRhcBzVbwDoe/PD08AoAA6fxXvWjSxy+dGhEaXoTHjkCbz/l6NxrK3JFyauDgU4K4MytsZ1HDiMgMW8hZXxszoICTTiQEcBBABAgAGBQJOTkelAAoJEKZP1bF62zmo79oH/1XDb29SYtWp+MTJTPFEwlWRiyRuDXy3wBd/BpwBRIWfWzMs1gnCjNjk0EVBVGa2grvy9JtxJKMd6l/PWXVucSt+U/+GO8rBkw14SdhqxaS2l14v6gyMeUrSbY3XfToGfwHC4sa/Thn8X4jFaQ2XN5dAIzJGU1s5JA0tjEzUwCnmrKmyMlXZaoQVrmORGjCuH0I0aAFkRS0UtnB9HPpxhGVbs24xXZQnZDNbUQeulFxS4uP3OLDBAeCHl+v4t/uotIad8v6JSO93vc1evIje6lguE81HHmJn9noxPItvOvSMb2yPsE8mH4cJHRTFNSEhPW6ghmlfWa9ZwiVX5igxcvaIRgQQEQIABgUCTk5b0gAKCRDs8OkLLBcgg1G+AKCnacLb/+W6cflirUIExgZdUJqoogCeNPVwXiHEIVqithAM1pdY/gcaQZmIRgQQEQIABgUCTk5fYQAKCRCpN2E5pSTFPnNWAJ9gUozyiS+9jf2rJvqmJSeWuCgVRwCcCUFhXRCpQO2YVa3l3WuB+rgKjsQ=" | base64 -d > /usr/share/keyrings/nginx.gpg \
    && set -x \
    && apt-get -y update && apt-get -y upgrade \
    && apt-get -y --no-install-recommends install \
        cron \
        logrotate \
        monitoring-plugins \
        supervisor \
        nginx \
        wallarm-node \
        nginx-module-wallarm \
        collectd-utils \
        curl \
        iptables \
        bsdmainutils \
        sudo \
        stunnel \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/* \
    && chown -R wallarm:wallarm /var/lib/wallarm-tarantool

RUN cp /usr/share/doc/nginx-module-wallarm/examples/wallarm-status.conf /etc/nginx/conf.d/ \
    && sed -i -e '/^worker_processes.*/a load_module modules/ngx_http_wallarm_module.so;' /etc/nginx/nginx.conf
COPY scripts/init /usr/local/bin/
COPY scripts/registernode_loop /usr/local/bin/

COPY conf/supervisord.conf /etc/supervisor/
COPY conf/supervisord.filtering.conf /etc/supervisor/supervisord.filtering.conf.example
COPY conf/supervisord.post-analytics.conf /etc/supervisor/supervisord.post-analytics.conf.example
COPY conf/stunnel.conf* /etc/stunnel/

COPY conf/logrotate.conf /etc/
COPY conf/default /etc/nginx/conf.d/default.conf
COPY conf/collectd.conf /etc/collectd/

EXPOSE 80 443

CMD ["/usr/local/bin/init"]
