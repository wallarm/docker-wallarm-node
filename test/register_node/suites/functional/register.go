//go:build functional

package functional

import (
	"context"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/ozontech/allure-go/pkg/framework/provider"

	"gl.wallarm.com/wallarm-node/aio-docker/test/register_node/shared"
)

// apiHost returns the default Wallarm API host used by positive/generic cases.
// Overridable via WALLARM_API_HOST so the same matrix can target any cloud
// (e.g. a devenv for NODE-7651 backward-compat runs) without forking cases.
// The two negative cases that intentionally use a *wrong* host
// (docs.wallarm.com, api.wallarm.com) stay hardcoded.
func apiHost() string {
	if h := os.Getenv("WALLARM_API_HOST"); h != "" {
		return h
	}
	return "audit.api.wallarm.com"
}

// nodeHostConfig returns a bind-mount HostConfig when WALLARM_GONODE_CONFIG
// points to a local yaml file — required for node-native-aio (go-node) images,
// which refuse to start without /opt/wallarm/etc/wallarm/go-node.yaml. Nil for
// meganode AiO images that don't need it.
func nodeHostConfig() *container.HostConfig {
	cfg := os.Getenv("WALLARM_GONODE_CONFIG")
	if cfg == "" {
		return nil
	}
	return &container.HostConfig{
		Binds: []string{cfg + ":/opt/wallarm/etc/wallarm/go-node.yaml:ro"},
	}
}

func (testSuite *RegisterSuite) TestRegisterNode(t provider.T) {
	t.Title("Register Node")
	t.Feature("Register Node")
	t.AllureID("19579")

	var testCases map[string]shared.RegisterNodeCases

	// Common allowed error patterns for all tests
	commonAllowedErrors := []string{
		"WALLARM:ACL: database is locked", // we are investigating this, but for now we ignore it
	}

	testCases = make(map[string]shared.RegisterNodeCases)
	for subscription, tokenTypes := range testSuite.tokens {
		for tokenType, token := range tokenTypes {
			caseName := "Register Node with " + subscription + " " + tokenType
			expectedResult := "\"node registration done\""
			allowedErrors := commonAllowedErrors
			if subscription == "EXPIRED" {
				expectedResult = "subscription limited"
				allowedErrors = append(allowedErrors, "subscription limited", "unable to open database file")
			}
			if subscription == "WAAP" {
				// WAAP subscription has API enforcement disabled, so we expect a warning about it
				allowedErrors = append(allowedErrors, "subscription limited for using api enforcement")
			}
			testCases[caseName] = shared.RegisterNodeCases{
				Token:                token,
				ExpectedResult:       expectedResult,
				TokenType:            tokenType,
				Subscription:         subscription,
				ApiHost:              apiHost(),
				ExpectFail:           false,
				AllowedErrorPatterns: allowedErrors,
			}
		}
	}

	// Negative cases live in register_negative.go (build tag !positive_only).
	// Build with -tags "functional,positive_only" to skip them — used for
	// go-node/native-aio runs where negative-case log signatures differ.
	appendNegativeCases(testCases, testSuite.tokens, commonAllowedErrors)

	for caseName, params := range testCases {
		paramsTest := params

		t.WithNewAsyncStep(caseName, func(stepCtx provider.StepCtx) {
			var (
				err         error
				out         io.ReadCloser
				success     bool
				logs        []byte
				strLogs     string
				envVars     []string
				resp        container.CreateResponse
				ReqInterval = 2 * time.Second
				ReqTimeout  = 80 * time.Second
				re          *regexp.Regexp
			)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			envVars = []string{"WALLARM_API_TOKEN=" + paramsTest.Token, "WALLARM_API_HOST=" + paramsTest.ApiHost}
			if paramsTest.TokenType == "api_token" {
				envVars = append(envVars, "WALLARM_LABELS=group=testWallarm")
			}
			stepCtx.Logf("Params: %v", paramsTest)

			resp, err = testSuite.dockerClient.ContainerCreate(ctx, &container.Config{
				Image: testSuite.imageName,
				Tty:   false,
				Env:   envVars,
			}, nodeHostConfig(), nil, nil, "")
			stepCtx.Require().NoError(err, "Error creating container")

			// Ensure container cleanup happens even if test fails
			defer func() {
				removeCtx := context.Background()
				removeErr := testSuite.dockerClient.ContainerRemove(
					removeCtx, resp.ID, container.RemoveOptions{RemoveVolumes: true, Force: true})
				if removeErr != nil {
					stepCtx.Logf("Warning: failed to remove container %s: %v", resp.ID, removeErr)
				}
			}()

			err = testSuite.dockerClient.ContainerStart(ctx, resp.ID, container.StartOptions{})
			stepCtx.Require().NoError(err, "Error starting container")

			success, err = shared.Poll(ReqInterval, ReqTimeout, func() (bool, error) {
				out, err = testSuite.dockerClient.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true, ShowStderr: true})
				if err != nil {
					stepCtx.Logf("Failed to get container logs: %v", err)
					return false, nil
				}
				defer out.Close()

				logs, err = io.ReadAll(out)
				if err != nil {
					stepCtx.Logf("Failed to read container logs: %v", err)
					return false, nil
				}
				strLogs = string(logs[:])

				// If we expect fail, we check if the expected message is in the logs
				if paramsTest.ExpectFail {
					if strings.Contains(strLogs, paramsTest.ExpectedResult) {
						return true, nil
					}
					// If we expect success, we check if the expected message is in the logs after wcli is in RUNNING state
				} else {
					if strings.Contains(strLogs, "wcli entered RUNNING state") {
						return true, nil
					}
				}
				return false, nil
			})

			stepCtx.WithNewAttachment("Container logs", "text/plain", logs)
			stepCtx.Require().True(success, "Error waiting node to be in the expected condition")
			stepCtx.Require().Contains(strLogs, paramsTest.ExpectedResult, "Error getting expected message: '%v' in log", paramsTest.ExpectedResult)

			// Check if there are any error messages in the logs (case insensitive)
			re = regexp.MustCompile(`(?i)error`)
			errorLines := []string{}

			// Split logs into lines and check each line
			for _, line := range strings.Split(strLogs, "\n") {
				if re.MatchString(line) {
					// Check if this error is in the allowed patterns
					isAllowed := false

					// If we have an expected error and this line contains it, it's allowed
					if paramsTest.ExpectedError != "" && strings.Contains(line, paramsTest.ExpectedError) {
						isAllowed = true
					}

					// Check against allowed error patterns
					for _, allowedPattern := range paramsTest.AllowedErrorPatterns {
						if matched, _ := regexp.MatchString(allowedPattern, line); matched {
							isAllowed = true
							break
						}
					}

					if !isAllowed {
						errorLines = append(errorLines, line)
					}
				}
			}

			if len(errorLines) > 0 {
				stepCtx.Logf("Found unexpected error messages in logs:\n%s", strings.Join(errorLines, "\n"))
			}
			stepCtx.Require().Empty(errorLines, "Found unexpected error messages in logs")
		})
	}
}

func (testSuite *RegisterSuite) TestUnRegisterNode(t provider.T) {
	t.Title("UnRegister Node")
	t.Feature("UnRegister Node")
	t.AllureID("19580")

	var (
		err             error
		resp            container.CreateResponse
		envVars         []string
		out             io.ReadCloser
		logs            []byte
		strLogs         string
		success         bool
		exec            types.IDResponse
		execInspect     container.ExecInspect
		wcliLog         []byte
		ReqInterval     = 2 * time.Second
		ReqTimeout      = 80 * time.Second
		execLog         types.HijackedResponse
		expectedMessage = "node unregistration done"
		re              *regexp.Regexp
		errors          []string
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	envVars = []string{
		"WALLARM_API_TOKEN=" + testSuite.tokens["NFR"]["node_token"],
		"WALLARM_API_HOST=" + apiHost(),
		"WALLARM_NODE_UNREGISTER=true",
	}

	resp, err = testSuite.dockerClient.ContainerCreate(ctx, &container.Config{
		Image: testSuite.imageName,
		Tty:   false,
		Env:   envVars,
	}, nodeHostConfig(), nil, nil, "")
	t.Require().NoError(err, "Error creating container")

	// Ensure container cleanup happens even if test fails
	defer func() {
		removeCtx := context.Background()
		removeErr := testSuite.dockerClient.ContainerRemove(
			removeCtx, resp.ID, container.RemoveOptions{RemoveVolumes: true, Force: true})
		if removeErr != nil {
			t.Logf("Warning: failed to remove container %s: %v", resp.ID, removeErr)
		}
	}()

	err = testSuite.dockerClient.ContainerStart(ctx, resp.ID, container.StartOptions{})
	t.Require().NoError(err, "Error starting container")

	// Wait for wcli to be in RUNNING state
	success, err = shared.Poll(ReqInterval, ReqTimeout, func() (bool, error) {
		out, err = testSuite.dockerClient.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true, ShowStderr: true})
		if err != nil {
			t.Logf("Failed to get container logs: %v", err)
			return false, nil
		}
		defer out.Close()

		logs, err = io.ReadAll(out)
		if err != nil {
			t.Logf("Failed to read container logs: %v", err)
			return false, nil
		}
		if strings.Contains(string(logs[:]), "wcli entered RUNNING state") {
			return true, nil
		}
		return false, nil
	})

	t.WithNewAttachment("Container logs", "text/plain", logs)
	t.Require().True(success, "Error waiting node to be in the expected condition")

	exec, err = testSuite.dockerClient.ContainerExecCreate(ctx, resp.ID, container.ExecOptions{
		Cmd:          []string{"bash", "-c", "kill $(pgrep wcli)"},
		AttachStderr: true,
		AttachStdout: true,
	})
	t.Require().NoError(err, "Error creating exec")

	execLog, err = testSuite.dockerClient.ContainerExecAttach(ctx, exec.ID, container.ExecStartOptions{})
	t.Require().NoError(err, "Error starting exec")

	wcliLog, err = io.ReadAll(execLog.Reader)
	if execLog.Reader != nil {
		execLog.Close()
	}
	t.WithNewAttachment("kill_wcli.log", "text/plain", wcliLog)

	execInspect, err = testSuite.dockerClient.ContainerExecInspect(ctx, exec.ID)
	t.Require().NoError(err, "Error executing kill to wcli")
	t.Require().Equal(execInspect.ExitCode, 0, "Error executing exec")

	success, err = shared.Poll(ReqInterval, ReqTimeout, func() (bool, error) {
		out, _, err = testSuite.dockerClient.CopyFromContainer(ctx, resp.ID, "/opt/wallarm/var/log/wallarm/wcli-out.log")
		if err != nil {
			t.Logf("Failed to get container logs: %v", err)
			return false, nil
		}
		defer out.Close()

		logs, err = io.ReadAll(out)
		strLogs = string(logs[:])

		if err != nil {
			t.Logf("Failed to read container logs: %v", err)
			return false, nil
		}
		if strings.Contains(strLogs, expectedMessage) {
			return true, nil
		}
		return false, nil
	})

	t.WithNewAttachment("wcli.log", "text/plain", logs)
	t.Require().True(success, "Error getting expected message: '%v' in log", expectedMessage)

	// Check if there are any errors in the logs
	re = regexp.MustCompile(`level\":\"error.*\n`)
	errors = re.FindAllString(strLogs, -1)
	t.Require().Empty(errors, "There are errors in the logs. See attachment for details")
}
