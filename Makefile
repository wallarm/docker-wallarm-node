-include .env

# Repo version pins
.EXPORT_ALL_VARIABLES:

AIO_VERSION       ?= 6.5.0

ALPINE_VERSION    = 3.22
NGINX_VERSION     = 1.28.0
GOMPLATE_VERISON  = 3.11.7
COMMIT_SHA        ?= git-$(shell git rev-parse --short HEAD)

NODE_DOCKER_IMAGE	?= node:test
IMAGE_NAME		?= node
IMAGE_TAG		?= $(shell git rev-parse --short HEAD)
IMAGE			?= $(NODE_DOCKER_IMAGE)
PUBLIC_REGISTRY ?= docker.io/wallarm/$(IMAGE_NAME)
DOCKERFILE		?= Dockerfile

SKIP_AIO_DOWNLOAD ?= false
BUILDX_ARGS = --push

ARCH ?= x86_64

DOCKER_ARCH ?= amd64
ifeq ($(ARCH),aarch64)
	DOCKER_ARCH := arm64
endif

DOCKER_SCOUT_ARGS ?= ""

PYTEST_WORKERS 	?= 10
PYTEST_PARAMS 	?= --allure-features=Node

.PHONY: docker-image-build docker-scout-scan docker-push docker-sign smoke-test single split test-register-node-ci test-register-node-local

docker-image-build:
	echo ${X_CI_BUILD_KIND}

	build-scripts/docker_build.sh
	@echo "IMAGE: ${IMAGE}"

docker-scout-scan:
	docker-scout cves "${NODE_DOCKER_IMAGE}" $(DOCKER_SCOUT_ARGS)

docker-push:
	docker buildx imagetools create -t ${PUBLIC_REGISTRY}:${IMAGE_TAG} ${NODE_DOCKER_IMAGE}

	# Will be fixed after NODE-6066
	if [ "$(X_CI_BUILD_KIND)" = "production" ] && ! echo "$(CI_COMMIT_TAG)" | grep -q '^sv-'; then \
		docker buildx imagetools create -t ${PUBLIC_REGISTRY}:latest ${PUBLIC_REGISTRY}:${IMAGE_TAG}; \
	fi

smoke-test: single split

single split:
	test/smoke_test.sh $@
