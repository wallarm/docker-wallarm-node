//go:build functional

package shared

type RegisterNodeCases struct {
	Token          string
	ExpectedResult string
	TokenType      string
	Subscription   string
	ApiHost        string
	ExpectFail     bool
	ExpectedError  string
}
