# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

Functional tests for Wallarm AIO Docker node registration/unregistration. Tests spin up Docker containers with the Wallarm node image and verify registration behavior across different subscription types (WAAP, NFR, FREE_TIER, AAS, EXPIRED) and token types (api_token, node_token), including negative cases.

## Tech Stack

- Go 1.23, build tag `functional`
- Test framework: [allure-go](https://github.com/ozontech/allure-go) (suite-based, with Allure reporting)
- Docker SDK for Go (`github.com/docker/docker/client`) — tests create/start/remove containers directly

## Commands

```bash
# Run tests locally (requires Docker, W_TEST_TOKENS, NODE_DOCKER_IMAGE env vars)
make test-register-node-local

# Equivalent manual command
go test -v -count=1 ./cmd/... -tags functional

# Build Docker image with test binary
make register-tests-build

# Run in CI (inside the test container, with allurectl)
make test-register-node-ci
```

## Required Environment Variables

- `W_TEST_TOKENS` — path to JSON file with tokens map (`map[subscription][tokenType]token`)
- `NODE_DOCKER_IMAGE` — full Docker image name to test against

## Architecture

- `cmd/functional_test.go` — entry point, wires `FunctionalSuite` → `RegisterSuite`
- `suites/functional/hooks.go` — `RegisterSuite` struct, `BeforeAll` (reads tokens, creates Docker client), `AfterAll` (cleanup)
- `suites/functional/register.go` — test methods: `TestRegisterNode` (data-driven, parallel async steps) and `TestUnRegisterNode`
- `shared/models.go` — `RegisterNodeCases` struct for test case parameters
- `shared/poll.go` — generic polling utility (interval + timeout)

Tests poll container logs waiting for `"wcli entered RUNNING state"` (success) or specific error messages (failure), then assert expected log content and check for unexpected errors.
