# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

Backward compatibility tests for Wallarm AIO Docker node. Verifies that a new Nginx Node version works correctly with a previous version of WCLI (wstore/post-analytics). Tests spin up a split-mode docker-compose setup (new node + old wstore with debug HTTP endpoint) and validate that attack detection, request processing, and response analysis work as expected across versions.

## Tech Stack

- Go 1.23, build tag `functional`
- Test framework: [allure-go](https://github.com/ozontech/allure-go) (suite-based, with Allure reporting)
- Docker Compose via `os/exec` — tests manage compose lifecycle directly
- Wstore debug HTTP endpoint on port 8989 (`WALLARM_WSTORE_DEBUG_HTTP_ENDPOINT`)

## Commands

```bash
# Run tests locally (requires Docker, NODE_IMAGE, WSTORE_IMAGE, WALLARM_API_TOKEN, WALLARM_API_HOST env vars)
make test-wstore-compat-local

# Equivalent manual command
go test -v -count=1 ./cmd/... -tags functional
```

## Required Environment Variables

- `NODE_DOCKER_IMAGE` — new node Docker image to test
- `WSTORE_IMAGE` — previous version node image (used as post-analytics/wstore)
- `WALLARM_API_TOKEN` — Wallarm API token for node registration
- `WALLARM_API_HOST` — Wallarm API host

## Architecture

- `cmd/functional_test.go` — entry point, wires `FunctionalSuite` → `WstoreCompatSuite`
- `suites/wstore/hooks.go` — `WstoreCompatSuite` struct, `BeforeAll` (compose up, waits for node to block attacks), `AfterAll` (collect logs, compose down)
- `suites/wstore/endpoints.go` — `TestDebugEndpoints`: two-phase test
  - Phase 1: sends SQLi attack, verifies attack-related debug endpoints (attacks, tags, uri, get, header, session_info)
  - Phase 2: sends clean request to `/api-discovery-test`, verifies response analysis endpoints (response_body, response_headers, response_points)
- `shared/poll.go` — generic polling utility (interval + timeout)

## Compose Setup

Uses `test/docker-compose.split_wstore.yaml`:
- `node` service: `${NODE_IMAGE}` with `POSTANALYTIC_ADDRESS=post-analytics`
- `post-analytics` service: `${WSTORE_IMAGE}` with `POSTANALYTIC_ONLY=true` and `WALLARM_WSTORE_DEBUG_HTTP_ENDPOINT=0.0.0.0:8989`

## Debug Endpoints Tested

All endpoints on wstore debug HTTP (`localhost:8989`):
- `/last_request/attacks` — detected attacks (sqli type, names)
- `/last_request/tags` — request tags (__blocked, final_wallarm_mode, libproton_version)
- `/last_request/uri` — request URI with random marker
- `/last_request/get` — GET parameters with marker validation
- `/last_request/header` — request headers (HOST, USER-AGENT)
- `/last_request/session_info` — session data (RuleID, Hash, Points)
- `/last_request/response_body` — captured response body (json_response)
- `/last_request/response_headers` — captured response headers (API-DISCOVERY-HEADER, CONTENT-TYPE)
- `/last_request/response_points` — response-level detection points
- `/last_request/post` — POST body (null for GET requests)
