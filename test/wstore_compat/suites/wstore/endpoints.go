//go:build functional

package wstore

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"gl.wallarm.com/wallarm-node/aio-docker/test/wstore_compat/shared"

	"github.com/ozontech/allure-go/pkg/framework/provider"
)

type endpointCheck struct {
	path     string
	validate func(body []byte, sCtx provider.StepCtx)
}

func (s *WstoreCompatSuite) TestDebugEndpoints(t provider.T) {
	t.Title("Wstore Debug Endpoints Compatibility")
	t.Feature("Wstore Compatibility")
	t.Epic("Wstore Compat")
	t.AllureID("NODE-7570")

	// --- Phase 1: Attack request ---
	attackMarker := fmt.Sprintf("wstoretest%d", time.Now().UnixNano())
	attackURL := fmt.Sprintf("%s/?%s=union+select+1", s.nodeURL, attackMarker)

	t.WithNewStep("Send attack request", func(sCtx provider.StepCtx) {
		resp, err := http.Get(attackURL)
		sCtx.Require().NoError(err, "failed to send attack request")
		defer resp.Body.Close()
		sCtx.Require().Equal(403, resp.StatusCode, "expected attack to be blocked")
		sCtx.Logf("Attack URL: %s", attackURL)
	})

	t.WithNewStep("Wait for wstore to process attack", func(sCtx provider.StepCtx) {
		ok, err := shared.Poll(2*time.Second, 60*time.Second, func() (bool, error) {
			resp, reqErr := http.Get(s.wstoreURL + "/last_request/attacks")
			if reqErr != nil {
				return false, nil
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				return false, nil
			}
			body, readErr := io.ReadAll(resp.Body)
			if readErr != nil {
				return false, nil
			}
			return len(body) > 4, nil
		})
		sCtx.Require().NoError(err)
		sCtx.Require().True(ok, "wstore did not process attack within timeout")
	})

	attackChecks := []endpointCheck{
		{
			path: "/last_request/session_info",
			validate: func(body []byte, sCtx provider.StepCtx) {
				var data map[string]interface{}
				sCtx.Require().NoError(json.Unmarshal(body, &data), "invalid JSON")
				sCtx.Assert().Contains(data, "RuleID", "missing RuleID")
				sCtx.Assert().Contains(data, "Hash", "missing Hash")
				sCtx.Assert().Contains(data, "Points", "missing Points")
			},
		},
		{
			path: "/last_request/attacks",
			validate: func(body []byte, sCtx provider.StepCtx) {
				var data map[string]interface{}
				sCtx.Require().NoError(json.Unmarshal(body, &data), "invalid JSON")
				sCtx.Assert().Contains(data, "attack_type_combined", "missing attack_type_combined")
				sCtx.Assert().Contains(data, "names", "missing names")
				names, ok := data["names"].([]interface{})
				sCtx.Require().True(ok, "names is not an array")
				sCtx.Assert().Contains(names, "sqli", "sqli not found in attack names")
			},
		},
		{
			path: "/last_request/tags",
			validate: func(body []byte, sCtx provider.StepCtx) {
				var data map[string]interface{}
				sCtx.Require().NoError(json.Unmarshal(body, &data), "invalid JSON")
				sCtx.Assert().Contains(data, "__attack_type", "missing __attack_type")
				sCtx.Assert().Contains(data, "__blocked", "missing __blocked")
				sCtx.Assert().Contains(data, "__request_id", "missing __request_id")
				sCtx.Assert().Contains(data, "final_wallarm_mode", "missing final_wallarm_mode")
				sCtx.Assert().Contains(data, "libproton_version", "missing libproton_version")
				sCtx.Assert().Contains(data, "lom_id", "missing lom_id")
				sCtx.Assert().Contains(data, "protondb_id", "missing protondb_id")
				if blocked, ok := data["__blocked"].(bool); ok {
					sCtx.Assert().True(blocked, "request was not blocked")
				}
			},
		},
		{
			path: "/last_request/post",
			validate: func(body []byte, sCtx provider.StepCtx) {
				sCtx.Assert().True(json.Valid(body), "invalid JSON")
			},
		},
		{
			path: "/last_request/uri",
			validate: func(body []byte, sCtx provider.StepCtx) {
				var data map[string]interface{}
				sCtx.Require().NoError(json.Unmarshal(body, &data), "invalid JSON")
				sCtx.Assert().Contains(data, "point", "missing point")
				sCtx.Assert().Contains(data, "value", "missing value")
				value, ok := data["value"].(string)
				sCtx.Require().True(ok, "value is not a string")
				sCtx.Assert().Contains(value, attackMarker, "URI does not contain marker")
			},
		},
		{
			path: "/last_request/get",
			validate: func(body []byte, sCtx provider.StepCtx) {
				var data map[string]interface{}
				sCtx.Require().NoError(json.Unmarshal(body, &data), "invalid JSON")
				sCtx.Assert().Contains(data, attackMarker, "GET params missing marker key")
				if paramRaw, ok := data[attackMarker].(map[string]interface{}); ok {
					sCtx.Assert().Contains(paramRaw, "point", "missing point in param")
					sCtx.Assert().Contains(paramRaw, "value", "missing value in param")
					if val, ok := paramRaw["value"].(string); ok {
						sCtx.Assert().Contains(val, "union select 1", "GET param value mismatch")
					}
				}
			},
		},
		{
			path: "/last_request/header",
			validate: func(body []byte, sCtx provider.StepCtx) {
				var data map[string]interface{}
				sCtx.Require().NoError(json.Unmarshal(body, &data), "invalid JSON")
				sCtx.Assert().Contains(data, "HOST", "missing HOST header")
				sCtx.Assert().Contains(data, "USER-AGENT", "missing USER-AGENT header")
			},
		},
	}

	runChecks(t, s.wstoreURL, attackChecks)

	// --- Phase 2: API Discovery request (response analysis) ---
	discoveryMarker := fmt.Sprintf("apidisc%d", time.Now().UnixNano())
	discoveryURL := fmt.Sprintf("%s/api-discovery-test/response-parameters/mytest?%s=testval", s.nodeURL, discoveryMarker)

	t.WithNewStep("Send API Discovery request", func(sCtx provider.StepCtx) {
		req, err := http.NewRequest("GET", discoveryURL, nil)
		sCtx.Require().NoError(err, "failed to create request")
		req.Header.Set("Custom-Id", "1")
		resp, err := http.DefaultClient.Do(req)
		sCtx.Require().NoError(err, "failed to send api-discovery request")
		defer resp.Body.Close()
		sCtx.Require().Equal(200, resp.StatusCode, "expected 200 from api-discovery-test")
		sCtx.Logf("Discovery URL: %s", discoveryURL)
	})

	t.WithNewStep("Wait for wstore to process api-discovery request", func(sCtx provider.StepCtx) {
		ok, err := shared.Poll(2*time.Second, 60*time.Second, func() (bool, error) {
			resp, reqErr := http.Get(s.wstoreURL + "/last_request/uri")
			if reqErr != nil {
				return false, nil
			}
			defer resp.Body.Close()
			body, readErr := io.ReadAll(resp.Body)
			if readErr != nil {
				return false, nil
			}
			return strings.Contains(string(body), discoveryMarker), nil
		})
		sCtx.Require().NoError(err)
		sCtx.Require().True(ok, "wstore did not process api-discovery request within timeout")
	})

	discoveryChecks := []endpointCheck{
		{
			// {"json_response": "Hello, I am json"} or similar captured body
			path: "/last_request/response_body",
			validate: func(body []byte, sCtx provider.StepCtx) {
				bodyStr := string(body)
				sCtx.Assert().True(len(bodyStr) > 4, "response_body is empty or null")
				sCtx.Assert().Contains(bodyStr, "json_response", "missing json_response in captured body")
				sCtx.Assert().Contains(bodyStr, "Hello, I am json", "missing expected response content")
			},
		},
		{
			// {"API-DISCOVERY-HEADER":{"point":[...],"value":"Hello, I am header!"},"CONTENT-TYPE":{...}}
			path: "/last_request/response_headers",
			validate: func(body []byte, sCtx provider.StepCtx) {
				var data map[string]interface{}
				sCtx.Require().NoError(json.Unmarshal(body, &data), "invalid JSON")
				sCtx.Assert().Contains(data, "API-DISCOVERY-HEADER", "missing API-DISCOVERY-HEADER")
				sCtx.Assert().Contains(data, "CONTENT-TYPE", "missing CONTENT-TYPE")
				if hdr, ok := data["API-DISCOVERY-HEADER"].(map[string]interface{}); ok {
					sCtx.Assert().Equal("Hello, I am header!", hdr["value"], "unexpected header value")
				}
				if ct, ok := data["CONTENT-TYPE"].(map[string]interface{}); ok {
					sCtx.Assert().Equal("application/json", ct["value"], "unexpected content-type")
				}
			},
		},
		{
			path: "/last_request/response_points",
			validate: func(body []byte, sCtx provider.StepCtx) {
				sCtx.Assert().True(json.Valid(body), "invalid JSON")
			},
		},
	}

	runChecks(t, s.wstoreURL, discoveryChecks)
}

func runChecks(t provider.T, wstoreURL string, checks []endpointCheck) {
	for _, check := range checks {
		check := check
		t.WithNewStep(fmt.Sprintf("Verify %s", check.path), func(sCtx provider.StepCtx) {
			resp, err := http.Get(wstoreURL + check.path)
			sCtx.Require().NoError(err, "GET %s failed", check.path)
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			sCtx.Require().NoError(err, "failed to read body from %s", check.path)

			sCtx.Assert().Equal(200, resp.StatusCode, "unexpected status for %s", check.path)

			check.validate(body, sCtx)

			attachName := strings.ReplaceAll(strings.TrimPrefix(check.path, "/"), "/", "_")
			sCtx.WithNewAttachment(attachName, "application/json", body)
		})
	}
}
