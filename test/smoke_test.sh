#!/bin/bash

function wait_for_node() {
  WAIT_TIMEOUT_SECONDS=60

  for i in $(seq 1 $WAIT_TIMEOUT_SECONDS); do
      printf '.'
      
      status="$(curl -I -o /dev/null -w %{http_code} 127.0.0.1:5000?sqli=union+select+1 || true)"
      if [[ "$status" -eq 403 ]]; then
          # If the attack was blocked, then wallarm-node started and enabled protection
          printf '\nINFO: wallarm-node started OK and enabled protection after %s seconds\n' "$i"
          return 0
      fi
      sleep 1

  done

  printf '\n'
  printf 'ERROR: wallarm-node failed to start within %s seconds\n' "$WAIT_TIMEOUT_SECONDS"
  return 1

}

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
      NODE_GROUP_NAME
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


# set -x
set -e
set -a

RED='\033[0;31m'
NC='\033[0m'

export NODE_GROUP_NAME="gitlab-docker-$(tr -dc A-Za-z0-9 </dev/urandom | head -c 12; echo)"
export WALLARM_LABELS="group=${NODE_GROUP_NAME}"

# check if all mandatory vars was defined
check_mandatory_vars

#single or split mode
NODE_MODE=$1
LOGS_DIR="${PWD}/test/logs/${NODE_MODE}"

COMPOSE_CMD="NODE_IMAGE=$IMAGE NODE_GROUP_NAME=$NODE_GROUP_NAME docker-compose -p $NODE_MODE -f test/docker-compose.$NODE_MODE.yaml"

echo "Staring Docker compose in ${NODE_MODE} mode using ${NODE_GROUP_NAME} group..."
eval "$COMPOSE_CMD up -d --wait --quiet-pull"

wait_for_node || get_logs_clean_and_exit

sleep 10 # wait node to export first attack

# trap for logs and exit
trap "get_logs_clean_and_exit" EXIT

GITLAB_VARS=()
while IFS= read -r line; do
  [[ -n "$line" ]] && GITLAB_VARS+=("$line")
done < <(printenv | grep -E '^(GITLAB_|ALLURE_)')

# set tests variables and run tests
EXEC_CMD=(
  env
  "${GITLAB_VARS[@]}"
  /usr/local/bin/test-entrypoint.sh
)

if [ "$ALLURE_UPLOAD_REPORT" = "true" ]; then
  EXEC_CMD+=(ci)
else
  EXEC_CMD+=(pytest ${PYTEST_PARAMS})
fi

echo "Running tests ..."
docker-compose -p $NODE_MODE -f test/docker-compose.$NODE_MODE.yaml exec pytest "${EXEC_CMD[@]}"
