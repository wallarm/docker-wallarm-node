# https://makefiletutorial.com/

# set default shell
SHELL=/bin/bash -o pipefail -o errexit

-include .env

DOCKERFILE := ./Dockerfile
REGISTRY   := docker.io/wallarm
TAG   	   ?= test
IMAGE 	   ?= $(REGISTRY)/node:$(TAG)

PYTEST_WORKERS   ?= 10
PYTEST_ARGS      ?= --allure-features=Node

COMPOSE_CMD  = NODE_IMAGE=$(IMAGE) docker-compose -p $@ -f test/docker-compose.$@.yaml
NODE_UUID    = $(COMPOSE_CMD) exec node cat /etc/wallarm/node.yaml | grep uuid | awk '{print $$2}'
PYTEST_CMD   = $(COMPOSE_CMD) exec -e NODE_UUID=$$($(NODE_UUID)) pytest pytest -n $(PYTEST_WORKERS) $(PYTEST_ARGS)

### Build routines
###
build:
	@docker build -t $(IMAGE) . --force-rm --no-cache --progress=plain

push rmi:
	@docker $@ $(IMAGE)

dive:
	@dive $(CONTROLLER_IMAGE)

.PHONY: build push rmi dive

### Smoke tests routines
###

smoke-test: single split

single split:
	$(COMPOSE_CMD) up -d --wait
	$(PYTEST_CMD)
	$(COMPOSE_CMD) down

.PHONY: smoke-test singe split
