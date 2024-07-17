#!/usr/bin/env bash

if [ "$1" = "test_config" ]; then
  /usr/sbin/nginx -t
else
  # Set shell mode to terminate if something goes wrong.
  set -euxo pipefail
  # Run Wallarm services. This could block for quite some time since we run register-node inside.
  /opt/wallarm/docker-init
  # You can add your custom logic here, but in the end nginx process must be run as well.
  /usr/sbin/nginx -g 'daemon off;'
  fi
fi
