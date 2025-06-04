-include .env

# Repo version pins
.EXPORT_ALL_VARIABLES:

AIO_VERSION       ?= 6.2.0-rc1

ALPINE_VERSION    = 3.20
NGINX_VERSION     = 1.26.3
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

test-register-node-ci:
	cd test/register_node/ ; \
	export ALLURE_LAUNCH_ID=$(allurectl launch create --format "ID" --no-header) ; \
	allurectl watch --results ./cmd/allure-results -- \
	go test -count=1 ./cmd/... -tags functional

test-register-node-local:
	cd test/register_node/ ; \
	rm -rf ./cmd/allure-results/ ; \
	go test -count=1 ./cmd/... -tags functional ; \
	allure serve ./cmd/allure-results/ -p 1234 -h 127.0.0.1
