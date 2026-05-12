.EXPORT_ALL_VARIABLES:
REGISTRY           ?= dkr.wallarm.com/wallarm-node/aio-docker
REGISTER_TESTS_IMG ?= $(REGISTRY)/register-tests
PLATFORM           ?= linux/amd64
REGISTER_TESTS_DOCKERFILE := ./Dockerfile
REGISTER_TESTS_CONTEXT    := $(dir $(REGISTER_TESTS_DOCKERFILE))
DOCKER_BUILDKIT ?= 1
ALLURE_RESULTS_PATH ?= /app/allure-results
IMAGE_TAG		?= $(shell git rev-parse --short HEAD)

.PHONY: register-tests-build register-tests-push

register-tests-build:
	docker buildx build \
		--platform $(PLATFORM) \
		--file $(REGISTER_TESTS_DOCKERFILE) \
		--tag  $(REGISTER_TESTS_IMG):$(IMAGE_TAG) \
		--load $(REGISTER_TESTS_CONTEXT)

register-tests-push: register-tests-build
	docker push $(REGISTER_TESTS_IMG):$(IMAGE_TAG)

register-tests-push-latest: register-tests-build
	docker tag $(REGISTER_TESTS_IMG):$(IMAGE_TAG) $(REGISTER_TESTS_IMG):latest
	docker push $(REGISTER_TESTS_IMG):latest

test-register-node-ci:
	export ALLURE_LAUNCH_ID=$(allurectl launch create --format "ID" --no-header) ; \
	export ALLURE_PROJECT_ID=3 ; \
	mkdir -p $(ALLURE_RESULTS_PATH) ; \
	allurectl watch --results $(ALLURE_RESULTS_PATH) -- \
	./register.test -test.v

test-register-node-local:
	rm -rf ./cmd/allure-results/ ; \
	go test -v -count=1 ./cmd/... -tags functional ; \
	allure serve ./cmd/allure-results/ -p 1234 -h 127.0.0.1
