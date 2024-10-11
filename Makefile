.EXPORT_ALL_VARIABLES:
# https://makefiletutorial.com/

# set default shell
SHELL=/bin/bash -o pipefail -o errexit

-include .env

AIO_VERSION       ?= 4.10.13
CONTAINER_VERSION ?= test
ALPINE_VERSION    = 3.20
NGINX_VERSION     = 1.26.1
GOMPLATE_VERISON  = 3.11.7
COMMIT_SHA        ?= git-$(shell git rev-parse --short HEAD)

REGISTRY     ?= docker.io/wallarm
IMAGE 	     ?= $(REGISTRY)/node:$(CONTAINER_VERSION)
IMAGE_LATEST := $(REGISTRY)/node:latest

NODE_VERSION    ?= $(shell echo $(AIO_VERSION) | awk -F'[-.]' '{print $$1"."$$2"."$$3}')

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

### Build routines
###
build:
	build-scripts/docker_build.sh

rmi:
	@docker rmi $(IMAGE)

push-latest:
	docker buildx imagetools create -t $(IMAGE_LATEST) $(IMAGE)

dive:
	@dive $(IMAGE)

smoke-test: single split

single split:
	test/smoke_test.sh $@

.PHONY: smoke-test singe split build push push-latest rmi dive
