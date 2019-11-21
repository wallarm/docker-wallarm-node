FROM debian:stretch
MAINTAINER Wallarm Support Team <support@wallarm.com>

ENV DEBIAN_FRONTEND noninteractive

RUN apt-get -y update \
    && apt-get -y install \
       apt-transport-https

COPY ssl/nginx-repo.key /etc/ssl/nginx/nginx-repo.key
COPY ssl/nginx-repo.crt /etc/ssl/nginx/nginx-repo.crt
COPY conf/sources.list /etc/apt/sources.list.d/wallarm-node.list
COPY conf/sources-nginx-plus.list /etc/apt/sources.list.d/nginx-plus.list
COPY conf/90nginx /etc/apt/apt.conf.d/90nginx

RUN printf -- "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" | base64 -d > /usr/share/keyrings/wallarm.gpg \
    && printf -- "mQENBE5OMmIBCAD+FPYKGriGGf7NqwKfWC83cBV01gabgVWQmZbMcFzeW+hMsgxHW6iimD0RsfZ9oEbfJCPG0CRSZ7ppq5pKamYs2+EJ8Q2ysOFHHwpGrA2C8zyNAs4IQxnZZIbETgcSwFtDun0XiqPwPZgyuXVm9PAbLZRbfBzm8wR/3SWygqZBBLdQk5TEfDR+Eny/M1RVR4xClECONF9UBB2ejFdI1LD45APbP2hsN/piFByU1t7yK2gpFyRt97WzGHn9MV5/TL7AmRPM4pcr3JacmtCnxXeCZ8nLqedoSuHFuhwyDnlAbu8I16O5XRrfzhrHRJFM1JnIiGmzZi6zBvH0ItfyX6ttABEBAAG0KW5naW54IHNpZ25pbmcga2V5IDxzaWduaW5nLWtleUBuZ2lueC5jb20+iQE+BBMBAgAoAhsDBgsJCAcDAgYVCAIJCgsEFgIDAQIeAQIXgAUCV2K1+AUJGB4fQQAKCRCr9b2Ce9m/YloaB/9XGrolkocm7l/tsVjaBQCteXKuwsm4XhCuAQ6YAwA1L1UheGOG/aa2xJvrXE8X32tgcTjrKoYoXWcdxaFjlXGTt6jV85qRguUzvMOxxSEM2Dn115etN9piPl0Zz+4rkx8+2vJGF+eMlruPXg/zd88NvyLq5gGHEsFRBMVufYmHtNfcp4okC1klWiRIRSdp4QY1wdrN1O+/oCTl8Bzy6hcHjLIq3aoumcLxMjtBoclc/5OTioLDwSDfVx7rWyfRhcBzVbwDoe/PD08AoAA6fxXvWjSxy+dGhEaXoTHjkCbz/l6NxrK3JFyauDgU4K4MytsZ1HDiMgMW8hZXxszoICTTiQEcBBABAgAGBQJOTkelAAoJEKZP1bF62zmo79oH/1XDb29SYtWp+MTJTPFEwlWRiyRuDXy3wBd/BpwBRIWfWzMs1gnCjNjk0EVBVGa2grvy9JtxJKMd6l/PWXVucSt+U/+GO8rBkw14SdhqxaS2l14v6gyMeUrSbY3XfToGfwHC4sa/Thn8X4jFaQ2XN5dAIzJGU1s5JA0tjEzUwCnmrKmyMlXZaoQVrmORGjCuH0I0aAFkRS0UtnB9HPpxhGVbs24xXZQnZDNbUQeulFxS4uP3OLDBAeCHl+v4t/uotIad8v6JSO93vc1evIje6lguE81HHmJn9noxPItvOvSMb2yPsE8mH4cJHRTFNSEhPW6ghmlfWa9ZwiVX5igxcvaIRgQQEQIABgUCTk5b0gAKCRDs8OkLLBcgg1G+AKCnacLb/+W6cflirUIExgZdUJqoogCeNPVwXiHEIVqithAM1pdY/gcaQZmIRgQQEQIABgUCTk5fYQAKCRCpN2E5pSTFPnNWAJ9gUozyiS+9jf2rJvqmJSeWuCgVRwCcCUFhXRCpQO2YVa3l3WuB+rgKjsQ=" | base64 -d > /usr/share/keyrings/nginx-plus.gpg \
    && apt-get -y update && apt-get -y upgrade \
    && apt-get -y --no-install-recommends install \
        cron \
        logrotate \
        monitoring-plugins \
        supervisor \
        nginx-plus \
        wallarm-node \
        nginx-plus-module-wallarm \
     && apt-get clean \
     && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/* \
     && chown -R wallarm:wallarm /var/lib/wallarm-tarantool \
     && sed -i -e 's|/var/log/wallarm/brute\.log|/var/log/wallarm/brute-detect\.log|' /etc/logrotate.d/wallarm-common \
     && sed -i -e 's|/usr/share/wallarm-common/syncnode|/usr/share/wallarm-common/syncnode -c /etc/wallarm-dist/node.yaml|' /etc/cron.d/wallarm-node-nginx \
     && rm -rf /etc/wallarm/triggers.d/ \
     && mkdir -p /etc/wallarm-dist/triggers.d \
     && mkdir -p /etc/nginx/sites-enabled/ \
     && sed -i -e '\@<Plugin syslog>@,\@</Plugin>@ s/^/#/' /etc/collectd/collectd.conf

COPY conf/node.yaml /etc/wallarm-dist/
COPY scripts/trigger /etc/wallarm-dist/triggers.d/nginx
COPY scripts/init /usr/local/bin/
COPY conf/supervisord.conf /etc/supervisor/
COPY conf/logrotate.conf /etc/
COPY conf/nginx.conf /etc/nginx/nginx.conf
COPY conf/proxy_params /etc/nginx/proxy_params
COPY conf/default /etc/nginx/sites-enabled/default

EXPOSE 80 443

CMD ["/usr/local/bin/init"]
