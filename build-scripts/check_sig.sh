#!/bin/sh
set -e

NGINX_PATH="/usr/sbin/nginx"
MODULE_PATH="/usr/lib/nginx/modules/ngx_http_wallarm_module.so"

NGINX_SIG=$(grep -E -ao '.,.,.,[01]{34}' "${NGINX_PATH}")
MODULE_SIG=$(grep -E -ao '.,.,.,[01]{34}' "${MODULE_PATH}")

if [ "${NGINX_SIG}" = "${MODULE_SIG}" ]; then
  echo "OK! Signature of nginx module match expectations from signature of nginx binary found in the base image"
else
  echo "Failure! The signature of module is mismatch: ${NGINX_SIG} / ${MODULE_SIG}"
  exit 1
fi
