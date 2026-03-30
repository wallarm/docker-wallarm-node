//go:build functional

package wstore

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"

	"gl.wallarm.com/wallarm-node/aio-docker/test/wstore_compat/shared"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
)

type WstoreCompatSuite struct {
	suite.Suite

	testDir     string
	projectName string
	nodeURL     string
	wstoreURL   string
	composeUp   bool
}

func (s *WstoreCompatSuite) BeforeAll(t provider.T) {
	for _, v := range []string{"NODE_DOCKER_IMAGE", "WSTORE_IMAGE", "WALLARM_API_TOKEN", "WALLARM_API_HOST"} {
		t.Require().NotEmpty(os.Getenv(v), v+" must be set")
	}

	s.testDir = "../.."
	s.projectName = fmt.Sprintf("wstore-compat-%d", time.Now().UnixNano())
	s.nodeURL = "http://127.0.0.1:5000"
	s.wstoreURL = "http://127.0.0.1:8989"

	if os.Getenv("NODE_GROUP_NAME") == "" {
		os.Setenv("NODE_GROUP_NAME", s.projectName)
		os.Setenv("WALLARM_LABELS", "group="+s.projectName)
	}

	cmd := s.composeCmd("up", "-d", "--wait", "--quiet-pull")
	output, err := cmd.CombinedOutput()
	t.Require().NoError(err, "docker compose up failed: %s", string(output))
	s.composeUp = true

	t.Log("Waiting for node to start blocking...")
	ok, err := shared.Poll(2*time.Second, 120*time.Second, func() (bool, error) {
		resp, reqErr := http.Get(s.nodeURL + "/?sqli=union+select+1")
		if reqErr != nil {
			return false, nil
		}
		defer resp.Body.Close()
		return resp.StatusCode == 403, nil
	})
	t.Require().NoError(err)
	t.Require().True(ok, "node did not start blocking within timeout")
	t.Log("Node is ready")
}

func (s *WstoreCompatSuite) AfterAll(t provider.T) {
	if !s.composeUp {
		return
	}

	cmd := s.composeCmd("logs", "--no-color")
	if output, err := cmd.CombinedOutput(); err == nil {
		t.WithNewAttachment("compose-logs", "text/plain", output)
	}

	cmd = s.composeCmd("down", "-v")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Logf("docker compose down failed: %s", string(output))
	}
}

func (s *WstoreCompatSuite) composeCmd(args ...string) *exec.Cmd {
	fullArgs := append([]string{"compose", "-p", s.projectName, "-f", "docker-compose.split_wstore.yaml"}, args...)
	cmd := exec.Command("docker", fullArgs...)
	cmd.Dir = s.testDir
	cmd.Env = os.Environ()
	return cmd
}
