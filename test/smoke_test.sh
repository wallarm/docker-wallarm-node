#!/bin/bash

function get_logs() {
    CONTAINER_NAME=$1

    [[ -d ${LOGS_DIR}/${CONTAINER_NAME} ]] || mkdir -p ${LOGS_DIR}/${CONTAINER_NAME}

    eval "$COMPOSE_CMD cp $CONTAINER_NAME:/opt/wallarm/var/log/wallarm/ ${LOGS_DIR}/$CONTAINER_NAME"
    eval "$COMPOSE_CMD logs --no-color $CONTAINER_NAME > ${LOGS_DIR}/$CONTAINER_NAME/$CONTAINER_NAME.log"
}

function get_logs_clean_and_exit() {

  get_logs node
  [[ $NODE_MODE == "split" ]] && get_logs post-analytics

  if [[ "${CI:-false}" == "true" ]]; then
          echo "We run in CI. Archiving logs ..."
          tar -czvf node-logs-${ARCH}-${NODE_MODE}.tar.gz -C ${LOGS_DIR} ./
  fi

  echo "Stopping Docker compose in ${NODE_MODE} mode ..."
  eval "$COMPOSE_CMD down"
}

function check_mandatory_vars() {

    declare -a mandatory
    declare -a allure_mandatory
    env_list=""

    # mandatory vars for smoke tests
    mandatory=(
      WALLARM_API_TOKEN
      WALLARM_API_HOST
      WALLARM_API_PRESET
      CLIENT_ID
      USER_TOKEN
      WEBHOOK_API_KEY
      WEBHOOK_UUID
    )

    for var in "${mandatory[@]}"; do
      if [[ -z "${!var:-}" ]]; then
        env_list+=" $var"
      fi
    done

    if [[ ! "${CI:-false}" == "true" ]]; then
      local_mandatory=(
        SMOKE_REGISTRY_TOKEN
        SMOKE_REGISTRY_SECRET
      )

      for var in "${local_mandatory[@]}"; do
        if [[ -z "${!var:-}" ]]; then
          env_list+=" $var"
        fi
      done
    fi

    if [[ "${ALLURE_UPLOAD_REPORT:-false}" == "true" ]]; then

      allure_mandatory=(
        ALLURE_TOKEN
        ALLURE_ENVIRONMENT_ARCH
        ALLURE_PROJECT_ID
      )

      for var in "${allure_mandatory[@]}"; do
        if [[ -z "${!var:-}" ]]; then
          env_list+=" $var"
        fi
      done
    fi

    if [[ -n "$env_list" ]]; then
      for var in ${env_list}; do
        echo -e "${RED}Environment variable $var must be set${NC}"
      done
      exit 1
    fi

}


set -x
set -e
set -a

RED='\033[0;31m'
NC='\033[0m'

# check if all mandatory vars was defined
check_mandatory_vars

#single or split mode
NODE_MODE=$1
LOGS_DIR="${PWD}/test/logs/${NODE_MODE}"

COMPOSE_CMD="NODE_IMAGE=$IMAGE docker-compose -p $NODE_MODE -f test/docker-compose.$NODE_MODE.yaml"

echo "Staring Docker compose in ${NODE_MODE} mode ..."
eval "$COMPOSE_CMD up -d --wait --quiet-pull"

# trap for logs and exit
trap "get_logs_clean_and_exit" EXIT

# set tests variables and run tests
GITHUB_VARS=$(env | awk -F '=' -v ORS==" " '/^GITHUB_/ {print "-e " $1 "=" $2}')
RAND_NUM=$(echo $RANDOM$RANDOM$RANDOM | cut -c 1-10)

RUN_TESTS="pytest pytest"
if [ "$ALLURE_UPLOAD_REPORT" = "true" ]; then
  RUN_TESTS="pytest allurectl watch --job-uid $RAND_NUM -- pytest"
fi

echo "Retrieving Wallarm Node UUID ..."
NODE_UUID=$(eval "$COMPOSE_CMD exec node cat /opt/wallarm/etc/wallarm/node.yaml | grep uuid | awk '{print \$2}'")
if [[ -z "${NODE_UUID}" ]]; then
  echo -e "${RED}Failed to retrieve Wallarm Node UUID${NC}"
  exit 1
fi

PYTEST_CMD="$COMPOSE_CMD exec $GITHUB_VARS -e NODE_UUID=$NODE_UUID $RUN_TESTS -n $PYTEST_WORKERS $PYTEST_ARGS"
echo "Running tests ..."
eval $PYTEST_CMD
