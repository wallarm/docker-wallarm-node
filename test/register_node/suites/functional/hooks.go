//go:build functional

package functional

import (
	"os"

	"encoding/json"
	"github.com/docker/docker/client"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
)

type RegisterSuite struct {
	suite.Suite

	dockerClient *client.Client
	imageName    string
	envVars      []string
	tokens       map[string]map[string]string
}

func (testSuite *RegisterSuite) BeforeAll(t provider.T) {
	var (
		err         error
		fileContent []byte
		tokensFile  string
		tokens      map[string]map[string]string
	)

	// Read tokens from file (see WALLARM_TEST_TOKENS in gitlab CI/CD variables)
	tokensFile = os.Getenv("WALLARM_TEST_TOKENS")
	t.Require().NotEmpty(tokensFile, "WALLARM_TEST_TOKENS is not set")

	fileContent, err = os.ReadFile(tokensFile)
	t.Require().NoError(err, "Error reading tokens file")

	err = json.Unmarshal(fileContent, &tokens)
	t.Require().NoError(err, "Error unmarshalling tokens file")

	testSuite.tokens = tokens

	testSuite.imageName = os.Getenv("NODE_DOCKER_IMAGE")
	t.Require().NotEmpty(testSuite.imageName, "NODE_DOCKER_IMAGE is not set")

	testSuite.dockerClient, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	t.Require().NoError(err, "Error creating docker client")
}

func (testSuite *RegisterSuite) BeforeEach(t provider.T) {
}

func (testSuite *RegisterSuite) AfterAll(t provider.T) {
}

func (testSuite *RegisterSuite) AfterEach(t provider.T) {
}
