//go:build functional && !positive_only

package functional

import (
	"gl.wallarm.com/wallarm-node/aio-docker/test/register_node/shared"
)

func appendNegativeCases(
	testCases map[string]shared.RegisterNodeCases,
	tokens map[string]map[string]string,
	commonAllowedErrors []string,
) {
	testCases["Register Node with WRONG token"] = shared.RegisterNodeCases{
		Token:                "BPwBS4jVGtLXNXx9nc",
		ExpectedResult:       "illegal base64 data",
		TokenType:            "node_token",
		Subscription:         "INVALID",
		ApiHost:              apiHost(),
		ExpectFail:           false,
		ExpectedError:        "illegal base64",
		AllowedErrorPatterns: append(commonAllowedErrors, "unable to open database file"),
	}
	testCases["Register Node with api_token no group"] = shared.RegisterNodeCases{
		Token:                tokens["AAS"]["api_token"],
		ExpectedResult:       "label \\\"group\\\" is required for this registration type",
		TokenType:            "node_token", // this is api token, but to test the error we use node_token
		Subscription:         "AAS",
		ApiHost:              apiHost(),
		ExpectFail:           false,
		ExpectedError:        "label \\\"group\\\" is required",
		AllowedErrorPatterns: append(commonAllowedErrors, "unable to open database file"),
	}
	testCases["Register Node with NFR wrong host"] = shared.RegisterNodeCases{
		Token:                tokens["NFR"]["node_token"],
		ExpectedResult:       "\"node registration done with error\"",
		TokenType:            "node_token",
		Subscription:         "NFR",
		ApiHost:              "docs.wallarm.com",
		ExpectFail:           false,
		ExpectedError:        "request []: not found",
		AllowedErrorPatterns: append(commonAllowedErrors, "unable to open database file"),
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
		AllowedErrorPatterns: append(commonAllowedErrors, "unable to open database file"),
	}
}