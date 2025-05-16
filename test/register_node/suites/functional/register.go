//go:build functional

package functional

import (
	"context"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/ozontech/allure-go/pkg/framework/provider"

	"gl.wallarm.com/wallarm-node/aio-docker/test/register_node/shared"
)

func (testSuite *RegisterSuite) TestRegisterNode(t provider.T) {
	t.Title("Register Node")
	t.Feature("Register Node")
	t.AllureID("19579")

	var testCases map[string]shared.RegisterNodeCases

	testCases = make(map[string]shared.RegisterNodeCases)
	for subscription, tokenTypes := range testSuite.tokens {
		for tokenType, token := range tokenTypes {
			caseName := "Register Node with " + subscription + " " + tokenType
			expectedResult := "\"node registration done\""
			if subscription == "EXPIRED" {
				expectedResult = "subscription limited"
			}
			testCases[caseName] = shared.RegisterNodeCases{
				Token:          token,
				ExpectedResult: expectedResult,
				TokenType:      tokenType,
				Subscription:   subscription,
				ApiHost:        "audit.api.wallarm.com",
				ExpectFail:     false,
			}
		}
	}

	// Negative cases
	testCases["Register Node with WRONG token"] = shared.RegisterNodeCases{
		Token:          "BPwBS4jVGtLXNXx9nc",
		ExpectedResult: "illegal base64 data",
		TokenType:      "node_token",
		Subscription:   "INVALID",
		ApiHost:        "audit.api.wallarm.com",
		ExpectFail:     false,
		ExpectedError:  "illegal base64",
	}
	testCases["Register Node with api_token no group"] = shared.RegisterNodeCases{
		Token:          testSuite.tokens["AAS"]["api_token"],
		ExpectedResult: "label \\\"group\\\" is required for this registration type",
		TokenType:      "node_token", // this is api token, but to test the error we use node_token
		Subscription:   "AAS",
		ApiHost:        "audit.api.wallarm.com",
		ExpectFail:     false,
		ExpectedError:  "label \\\"group\\\" is required",
	}
	testCases["Register Node with NFR wrong host"] = shared.RegisterNodeCases{
		Token:          testSuite.tokens["NFR"]["node_token"],
		ExpectedResult: "\"node registration done with error\"",
		TokenType:      "node_token",
		Subscription:   "NFR",
		ApiHost:        "docs.wallarm.com",
		ExpectFail:     false,
		ExpectedError:  "request []: not found",
	}
	testCases["Register Node with no token"] = shared.RegisterNodeCases{
		Token:          "",
		ExpectedResult: "no WALLARM_API_TOKEN",
		TokenType:      "node_token",
		Subscription:   "NFR",
		ApiHost:        "audit.api.wallarm.com",
		ExpectFail:     true,
		ExpectedError:  "no private key",
	}
	testCases["Register Node with token from other host"] = shared.RegisterNodeCases{
		Token:          testSuite.tokens["NFR"]["node_token"],
		ExpectedResult: "access denied",
		TokenType:      "node_token",
		Subscription:   "NFR",
		ApiHost:        "api.wallarm.com",
		ExpectFail:     false,
		ExpectedError:  "access denied",
	}

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
				errors      []string
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
			}, nil, nil, nil, "")
			stepCtx.Require().NoError(err, "Error creating container")

			err = testSuite.dockerClient.ContainerStart(ctx, resp.ID, container.StartOptions{})
			stepCtx.Require().NoError(err, "Error starting container")

			success, err = shared.Poll(ReqInterval, ReqTimeout, func() (bool, error) {
				out, err = testSuite.dockerClient.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true, ShowStderr: true})
				if err != nil {
					stepCtx.Logf("Failed to get container logs: %v", err)
					return false, nil
				}
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

			stepCtx.WithNewStep("Delete container", func(subStepCtx provider.StepCtx) {

				ctxNew := context.Background()
				subStepCtx.Assert().
					NoError(testSuite.dockerClient.ContainerRemove(
						ctxNew, resp.ID, container.RemoveOptions{RemoveVolumes: true, Force: true}),
						"Error stopping image")
			})

			stepCtx.WithNewAttachment("Container logs", "text/plain", logs)
			stepCtx.Require().True(success, "Error waiting node to be in the expected condition")
			stepCtx.Require().Contains(strLogs, paramsTest.ExpectedResult, "Error getting expected message: '%v' in log", paramsTest.ExpectedResult)

			// Check if there are any unexpected errors in the logs
			re = regexp.MustCompile(`level\":\"error.*\n`)
			errors = re.FindAllString(strLogs, -1)

			for _, logError := range errors {
				stepCtx.Require().Contains(logError, paramsTest.ExpectedError, "Error should match the expected: '%v'", logError)
			}
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
		re		        *regexp.Regexp
		errors   	    []string
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	envVars = []string{
		"WALLARM_API_TOKEN=" + testSuite.tokens["NFR"]["node_token"],
		"WALLARM_API_HOST=audit.api.wallarm.com",
		"WALLARM_NODE_UNREGISTER=true",
	}

	resp, err = testSuite.dockerClient.ContainerCreate(ctx, &container.Config{
		Image: testSuite.imageName,
		Tty:   false,
		Env:   envVars,
	}, nil, nil, nil, "")
	t.Require().NoError(err, "Error creating container")

	err = testSuite.dockerClient.ContainerStart(ctx, resp.ID, container.StartOptions{})
	t.Require().NoError(err, "Error starting container")

	// Wait for wcli to be in RUNNING state
	success, err = shared.Poll(ReqInterval, ReqTimeout, func() (bool, error) {
		out, err = testSuite.dockerClient.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true, ShowStderr: true})
		if err != nil {
			t.Logf("Failed to get container logs: %v", err)
			return false, nil
		}
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

	t.WithNewStep("Delete container", func(subStepCtx provider.StepCtx) {
		ctxNew := context.Background()
		subStepCtx.Assert().
			NoError(testSuite.dockerClient.ContainerRemove(
				ctxNew, resp.ID, container.RemoveOptions{RemoveVolumes: true, Force: true}),
				"Error stopping image")
	})
	t.Require().True(success, "Error getting expected message: '%v' in log", expectedMessage)

	// Check if there are any errors in the logs
	re = regexp.MustCompile(`level\":\"error.*\n`)
	errors = re.FindAllString(strLogs, -1)
	t.Require().Empty(errors, "There are errors in the logs. See attachment for details")
}
