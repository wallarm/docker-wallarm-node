#!/bin/sh

set -e

prepare_dirs() {
  rm -f /etc/wallarm-dist/triggers.d/nginx
  touch /etc/crontab /etc/cron.d/*

  mkdir -p /run/supervisor
}

register_node() {
  WALLARM_API_USE_SSL="${WALLARM_API_USE_SSL:-true}"
  WALLARM_API_CA_VERIFY="${WALLARM_API_CA_VERIFY:-true}"
  /usr/share/wallarm-common/synccloud --one-time
}

if [ "x${SLAB_ALLOC_ARENA}" = 'x' ]; then
  if [ -n "$TARANTOOL_MEMORY_GB" ]; then
    SLAB_ALLOC_ARENA=$TARANTOOL_MEMORY_GB
    export SLAB_ALLOC_ARENA
  fi
fi

prepare_dirs
register_node

exec /usr/bin/supervisord -c /etc/supervisor/supervisord-tarantool.conf
