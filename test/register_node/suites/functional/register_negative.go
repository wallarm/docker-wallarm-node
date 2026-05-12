//go:build functional && !positive_only

package functional

import (
	"gl.wallarm.com/wallarm-node/aio-docker/test/register_node/shared"
)

// appendNegativeCases registers the five intentionally-failing test cases.
// All five share the same allow-list (negativeAllowedErrors, defined in
// register.go) — Cloud-rejection markers (deployCloudNode: not found,
// init stage failed, fallback mode, wstore metrics-exporter refresh
// failure) are expected on every negative path and stay scoped to this
// file so positive cases don't silently swallow the same lines.
//
// The third parameter is kept for backward compatibility with the
// signature on main; it is unused here because we now resolve the
// allow-list through the package-scope variable.
func appendNegativeCases(
	testCases map[string]shared.RegisterNodeCases,
	tokens map[string]map[string]string,
	_ []string,
) {
	testCases["Register Node with WRONG token"] = shared.RegisterNodeCases{
		Token:                "BPwBS4jVGtLXNXx9nc",
		ExpectedResult:       "illegal base64 data",
		TokenType:            "node_token",
		Subscription:         "INVALID",
		ApiHost:              apiHost(),
		ExpectFail:           false,
		ExpectedError:        "illegal base64",
		AllowedErrorPatterns: negativeAllowedErrors,
	}
	testCases["Register Node with api_token no group"] = shared.RegisterNodeCases{
		Token:                tokens["AAS"]["api_token"],
		ExpectedResult:       "label \\\"group\\\" is required for this registration type",
		TokenType:            "node_token", // this is api token, but to test the error we use node_token
		Subscription:         "AAS",
		ApiHost:              apiHost(),
		ExpectFail:           false,
		ExpectedError:        "label \\\"group\\\" is required",
		AllowedErrorPatterns: negativeAllowedErrors,
	}
	testCases["Register Node with NFR wrong host"] = shared.RegisterNodeCases{
		Token:                tokens["NFR"]["node_token"],
		ExpectedResult:       "\"node registration done with error\"",
		TokenType:            "node_token",
		Subscription:         "NFR",
		ApiHost:              "docs.wallarm.com",
		ExpectFail:           false,
		ExpectedError:        "request []: not found",
		AllowedErrorPatterns: negativeAllowedErrors,
	}
	testCases["Register Node with no token"] = shared.RegisterNodeCases{
		Token:                "",
		ExpectedResult:       "no WALLARM_API_TOKEN",
		TokenType:            "node_token",
		Subscription:         "NFR",
		ApiHost:              apiHost(),
		ExpectFail:           true,
		ExpectedError:        "no private key",
		AllowedErrorPatterns: []string{},
	}
	testCases["Register Node with token from other host"] = shared.RegisterNodeCases{
		Token:                tokens["NFR"]["node_token"],
		ExpectedResult:       "access denied",
		TokenType:            "node_token",
		Subscription:         "NFR",
		ApiHost:              "api.wallarm.com",
		ExpectFail:           false,
		ExpectedError:        "access denied",
		AllowedErrorPatterns: negativeAllowedErrors,
	}
}