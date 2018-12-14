#!/bin/sh

set -e

prepare_dirs() {
  touch /etc/crontab /etc/cron.d/*

  mkdir -p /run/supervisor
}

register_node() {
  WALLARM_API_USE_SSL="${WALLARM_API_USE_SSL:-true}"
  WALLARM_API_CA_VERIFY="${WALLARM_API_CA_VERIFY:-true}"
  /usr/share/wallarm-common/synccloud --one-time
}

configure_nginx() {
  [ -n "$NGINX_BACKEND" ] || return 0

  if [ "$NGINX_BACKEND" = "${NGINX_BACKEND#http://}" -a \
       "$NGINX_BACKEND" = "${NGINX_BACKEND#https://}" ];
  then
    sed -i -e "s#proxy_pass .*#proxy_pass http://$NGINX_BACKEND;#" \
      /etc/nginx/sites-enabled/default
  else
    sed -i -e "s#proxy_pass .*#proxy_pass $NGINX_BACKEND;#" \
      /etc/nginx/sites-enabled/default
  fi

  sed -i -e "s@# wallarm_mode .*@wallarm_mode ${WALLARM_MODE:-monitoring};@" \
    /etc/nginx/sites-enabled/default

  if [ -n "$WALLARM_INSTANCE" ]; then
    sed -i -e "s@# wallarm_instance .*@wallarm_instance ${WALLARM_INSTANCE};@" \
      /etc/nginx/sites-enabled/default
  fi
}

configure_tarantool_upstream() {
  TARANTOOL_HOST="${TARANTOOL_HOST:-127.0.0.1}"
  TARANTOOL_PORT="${TARANTOOL_PORT:-3313}"

  cat <<EOF >/etc/nginx/conf.d/tarantool-upstream.conf
upstream wallarm_tarantool {
	server $TARANTOOL_HOST:$TARANTOOL_PORT max_fails=3 fail_timeout=1;
	keepalive 1;
}

wallarm_tarantool_upstream wallarm_tarantool;
EOF
}

prepare_dirs
register_node
configure_tarantool_upstream
configure_nginx

exec /usr/bin/supervisord -c /etc/supervisor/supervisord-nginx.conf
