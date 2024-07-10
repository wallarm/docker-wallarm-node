.EXPORT_ALL_VARIABLES:
# https://makefiletutorial.com/

# set default shell
SHELL=/bin/bash -o pipefail -o errexit

-include .env

AIO_VERSION       ?= 4.10.7
CONTAINER_VERSION ?= test
ALPINE_VERSION    = 3.20
NGINX_VERSION     = 1.26.1
WLRM_FOLDER       = stable-$(shell echo ${NGINX_VERSION} | sed 's/\.//g')
GOMPLATE_VERISON  = 3.11.7
COMMIT_SHA        ?= git-$(shell git rev-parse --short HEAD)

REGISTRY     ?= docker.io/wallarm
IMAGE 	     ?= $(REGISTRY)/node:$(CONTAINER_VERSION)
IMAGE_LATEST := $(REGISTRY)/node:latest

RAND_NUM        := $(shell echo $$RANDOM$$RANDOM$$RANDOM | cut -c 1-10)
NODE_VERSION    ?= $(shell echo $(AIO_VERSION) | awk -F'[-.]' '{print $$1"."$$2"."$$3}')
COMPOSE_CMD     = NODE_IMAGE=$(IMAGE) docker-compose -p $@ -f test/docker-compose.$@.yaml
NODE_UUID_CMD   = $(COMPOSE_CMD) exec node cat /opt/wallarm/etc/wallarm/node.yaml | grep uuid | awk '{print $$2}'
NODE_UUID       = $(shell $(NODE_UUID_CMD))
GITHUB_VARS_CMD = env | awk -F '=' '/^GITHUB_/ {print "-e " $$1 "=" $$2}'
GITHUB_VARS     = $(shell $(GITHUB_VARS_CMD))
RUN_TESTS       := $(shell [ "$$ALLURE_UPLOAD_REPORT" = "true" ] && \
                     echo "pytest allurectl watch --job-uid $(RAND_NUM) -- pytest" || \
                     echo "pytest pytest")
PYTEST_CMD = $(COMPOSE_CMD) exec $(GITHUB_VARS) -e NODE_UUID=$$($(NODE_UUID_CMD)) \
             $(RUN_TESTS) -n $(PYTEST_WORKERS) $(PYTEST_ARGS)

### Variables required to run test
.EXPORT_ALL_VARIABLES:
ALLURE_ENDPOINT       ?= https://allure.wallarm.com
ALLURE_PROJECT_ID     ?= 10
WALLARM_API_HOST      ?= api.wallarm.com
WALLARM_API_CA_VERIFY ?= True
CLIENT_ID             ?= 5
PYTEST_WORKERS        ?= 10
PYTEST_ARGS           ?= --allure-features=Node
TEST_RS               ?= false
WALLARM_LABELS		  ?='group=defaultDockerNode'

# Single-platform for local, multi-platform for CI
ifndef CI
	ARCH ?=$(shell uname -m)
	ifeq ($(ARCH), x86_64)
		PLATFORMS?=linux/amd64
		ARCHS?=x86_64
	else ifeq ($(ARCH), arm64)
		PLATFORMS?=linux/arm64
		ARCHS?=aarch64
	else ifeq ($(ARCH), amd64)
		PLATFORMS?=linux/amd64
		ARCHS?=x86_64
	else
		$(error Unsupported architecture "$(ARCH)")
	endif
	BUILDX_ARGS?=--load
else
	PLATFORMS?=linux/amd64,linux/aarch64
	ARCHS?=x86_64 aarch64
	BUILDX_ARGS?=--push
endif

### Build routines
###
.PHONY: build
build: setup_buildx
	@echo "AIO version: $(AIO_VERSION)"

ifndef SKIP_AIO_DOWNLOAD
	$(foreach ARCH,$(ARCHS), ARCH=$(ARCH) build-scripts/get_dependencies.sh ;)
endif

	$(foreach ARCH,$(ARCHS), ARCH=$(ARCH) build-scripts/apply_fixes.sh ;)

	docker buildx build \
		--platform $(PLATFORMS) -f Dockerfile \
		--build-arg CONTAINER_VERSION="$(CONTAINER_VERSION)" \
		--build-arg GOMPLATE_VERISON="$(GOMPLATE_VERISON)" \
		--build-arg ALPINE_VERSION="$(ALPINE_VERSION)" \
		--build-arg NGINX_VERSION="$(NGINX_VERSION)" \
		--build-arg AIO_VERSION="$(AIO_VERSION)" \
		--build-arg WLRM_FOLDER="$(WLRM_FOLDER)" \
		--build-arg COMMIT_SHA="$(COMMIT_SHA)" \
		--no-cache \
		-t $(IMAGE) $(BUILDX_ARGS) .

setup_buildx:
	docker buildx rm multi-arch || true
	docker buildx create \
		--name multi-arch \
		--platform linux/amd64,linux/arm64 \
		--driver docker-container \
		--use

rmi:
	@docker rmi $(IMAGE)

push-latest:
	docker buildx imagetools create -t $(IMAGE_LATEST) $(IMAGE)

dive:
	@dive $(IMAGE)

.PHONY: build push push-latest rmi dive

### Smoke tests routines
###

smoke-test: single split

single split:
	$(COMPOSE_CMD) up -d --wait --quiet-pull
	$(PYTEST_CMD)
	$(COMPOSE_CMD) down

.PHONY: smoke-test singe split
