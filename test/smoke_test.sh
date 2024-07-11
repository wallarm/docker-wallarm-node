#!/bin/bash

function get_logs() {
    LOGS_DIR=$1
    CONTAINER_NAME=$2

    [[ -d ${LOGS_DIR}/${CONTAINER_NAME} ]] || mkdir -p ${LOGS_DIR}/${CONTAINER_NAME}

    /bin/bash -c "$COMPOSE_CMD cp $CONTAINER_NAME:/opt/wallarm/var/log/wallarm/ ${LOGS_DIR}/$CONTAINER_NAME"
    /bin/bash -c "$COMPOSE_CMD logs --no-color $CONTAINER_NAME > ${LOGS_DIR}/$CONTAINER_NAME/$CONTAINER_NAME.log"
}

set -x
set -e
set -a

NODE_MODE=$1 #single or split mode
LOGS_DIR="${PWD}/test/logs/${NODE_MODE}"

COMPOSE_CMD="NODE_IMAGE=$IMAGE docker-compose -p $NODE_MODE -f test/docker-compose.$NODE_MODE.yaml"

# Up compose
/bin/bash -c "$COMPOSE_CMD up -d --wait --quiet-pull"

# set tests variables and run tests
GITHUB_VARS=$(env | awk -F '=' -v ORS==" " '/^GITHUB_/ {print "-e " $1 "=" $2}')
RAND_NUM=$(echo $RANDOM$RANDOM$RANDOM | cut -c 1-10)

RUN_TESTS="pytest pytest"
if [ "$ALLURE_UPLOAD_REPORT" = "true" ]; then
  RUN_TESTS="pytest allurectl watch --job-uid $RAND_NUM -- pytest"
fi

NODE_UUID=$(/bin/bash -c "$COMPOSE_CMD exec node cat /opt/wallarm/etc/wallarm/node.yaml | grep uuid | awk '{print \$2}'")
PYTEST_CMD="$COMPOSE_CMD exec $GITHUB_VARS -e NODE_UUID=$NODE_UUID $RUN_TESTS -n $PYTEST_WORKERS $PYTEST_ARGS"

set +e #we need to get logs anyway, even if the tests were not successful.
/bin/bash -c "$PYTEST_CMD"
TESTS_STATUS=$?

set -e

# Get logs from containers
get_logs $LOGS_DIR node

[[ $NODE_MODE == "split" ]] && get_logs $LOGS_DIR post-analytics

if [[ "${CI:-false}" == "true" ]]; then
        echo "We run in CI. Archiving logs ..."
        tar -czvf node-logs-${ARCH}-${NODE_MODE}.tar.gz -C ${LOGS_DIR} ./
fi

# down compose
/bin/bash -c "$COMPOSE_CMD down"

# exit with test exit code
exit $TESTS_STATUS
