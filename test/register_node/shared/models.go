//go:build functional

package shared

// RegisterNodeCases represents a test case for node registration/unregistration tests.
// It defines the parameters and expected outcomes for testing Wallarm node registration
// with different token types, subscriptions, and API configurations.
type RegisterNodeCases struct {
	// Token is the Wallarm API token used for authentication (can be api_token or node_token)
	Token string

	// ExpectedResult is the expected message in container logs indicating success or specific error
	ExpectedResult string

	// TokenType specifies the type of token being used ("api_token" or "node_token")
	TokenType string

	// Subscription indicates the subscription type (e.g., "WAAP", "NFR", "FREE_TIER", "AAS", "EXPIRED")
	Subscription string

	// ApiHost is the Wallarm API host URL to connect to
	ApiHost string

	// ExpectFail indicates whether the test is expected to fail before reaching RUNNING state
	ExpectFail bool

	// ExpectedError is the expected error message pattern in logs (empty if no specific error expected)
	ExpectedError string

	// AllowedErrorPatterns is a list of regex patterns for allowed error messages that should be ignored
	AllowedErrorPatterns []string
}
