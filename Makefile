.PHONY: all
all: build

.PHONY: build
build:
	docker build --pull --build-arg WALLARM_NODE_MODE -t aio-docker .
