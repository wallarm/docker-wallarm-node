#!/usr/bin/env bash

if [ "$1" = "test_config" ]; then
  /usr/sbin/nginx -t
else
  # Set shell mode to terminate if something goes wrong.
  set -euxo pipefail
  # Run Wallarm services. This could block for quite some time since we run register-node inside.
  /opt/wallarm/docker-init
  # Depending on WALLARM_NODE_MODE, we either start nginx or do not start it and sleep indefinitely.
  if [ "$WALLARM_NODE_MODE" = "postanalytics" ]; then
    # Postanalytics mode doesn't have nginx, so just sleep indefinitely.
    sleep infinity
  else
    # Other modes must run nginx.
    /usr/sbin/nginx -g 'daemon off;'
  fi
fi
