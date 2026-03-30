ALLURE_RESULTS_PATH ?= ./cmd/allure-results

.PHONY: test-wstore-compat-local test-wstore-compat-ci

test-wstore-compat-ci:
	export ALLURE_LAUNCH_ID=$(allurectl launch create --format "ID" --no-header) ; \
	export ALLURE_PROJECT_ID=3 ; \
	mkdir -p $(ALLURE_RESULTS_PATH) ; \
	allurectl watch --results $(ALLURE_RESULTS_PATH) -- \
	go test -v -count=1 ./cmd/... -tags functional

test-wstore-compat-local:
	rm -rf $(ALLURE_RESULTS_PATH) ; \
	go test -v -count=1 ./cmd/... -tags functional ; \
	allure serve $(ALLURE_RESULTS_PATH) -p 1234 -h 127.0.0.1
