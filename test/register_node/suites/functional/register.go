//go:build functional

package functional

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
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

// readyCheckViaExec assesses wd readiness INSIDE the target container by
// running `wd ctl -j status`. Avoids both inter-container networking
// (DinD-runner setup → test and wd are on different daemons / bridges)
// and any dependency on the health HTTP listener / bash / curl. wd ctl
// talks to wd's Unix socket — the same path operators use to inspect
// the daemon, so the test exercises the real operational primitive.
//
// Ready when:
//   - the daemon answered ok=true (ctl socket is responsive);
//   - init_state is `done`, `fallback`, `synced`, or empty (legacy);
//   - every reported process is in state="running".
func readyCheckViaExec(ctx context.Context, cli *client.Client, containerID string) (ready bool, body string, err error) {
	exec, err := cli.ContainerExecCreate(ctx, containerID, container.ExecOptions{
		Cmd:          []string{"/opt/wallarm/usr/bin/wd", "ctl", "-j", "status"},
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		return false, "", fmt.Errorf("exec create: %w", err)
	}

	attach, err := cli.ContainerExecAttach(ctx, exec.ID, container.ExecStartOptions{})
	if err != nil {
		return false, "", fmt.Errorf("exec attach: %w", err)
	}
	defer attach.Close()

	var buf bytes.Buffer
	if _, err := stdcopy.StdCopy(&buf, &buf, attach.Reader); err != nil && err != io.EOF {
		return false, "", fmt.Errorf("exec read: %w", err)
	}
	out := buf.String()

	insp, err := cli.ContainerExecInspect(ctx, exec.ID)
	if err != nil {
		return false, out, fmt.Errorf("exec inspect: %w", err)
	}
	if insp.ExitCode != 0 {
		// Most common failure here: ctl socket not yet bound (wd still
		// in the early init phase). Treat as not-ready, retry.
		return false, out, fmt.Errorf("wd ctl exit %d", insp.ExitCode)
	}

	var resp struct {
		OK        bool   `json:"ok"`
		Error     string `json:"error,omitempty"`
		InitState string `json:"init_state,omitempty"`
		Status    []struct {
			Name          string `json:"name"`
			State         string `json:"state"`
			BlockedByInit bool   `json:"blocked_by_init,omitempty"`
		} `json:"status,omitempty"`
	}
	if err := json.Unmarshal([]byte(out), &resp); err != nil {
		return false, out, fmt.Errorf("decode: %w", err)
	}
	if !resp.OK {
		return false, out, fmt.Errorf("wd ctl ok=false: %s", resp.Error)
	}
	switch resp.InitState {
	case "done", "fallback", "synced", "":
		// past init pipeline
	default:
		return false, out, fmt.Errorf("init_state=%q", resp.InitState)
	}
	for _, p := range resp.Status {
		if p.BlockedByInit {
			// fallback mode: this process is intentionally not started
			continue
		}
		if p.State != "running" {
			return false, out, fmt.Errorf("process %q state=%q", p.Name, p.State)
		}
	}
	return true, out, nil
}

// containerIP extracts an IP address from a Docker ContainerInspect result.
// The legacy NetworkSettings.IPAddress field is only populated for the default
// `bridge` network on a host-bridge driver. Containers attached to a custom
// network (compose-default, user-defined bridge, DinD) leave that field empty
// and instead populate NetworkSettings.Networks[name].IPAddress. Probing both
// keeps the test robust across local Docker, CI DinD, and compose setups.
//
// Returns the first non-empty IP found; empty string if the container has no
// usable IP yet (caller should keep polling).
func containerIP(inspect types.ContainerJSON) string {
	if inspect.NetworkSettings == nil {
		return ""
	}
	if inspect.NetworkSettings.IPAddress != "" {
		return inspect.NetworkSettings.IPAddress
	}
	for _, network := range inspect.NetworkSettings.Networks {
		if network != nil && network.IPAddress != "" {
			return network.IPAddress
		}
	}
	return ""
}

// commonAllowedErrors is the set of error patterns ignored by all tests in this suite.
var commonAllowedErrors = []string{
	"WALLARM:ACL: database is locked", // we are investigating this, but for now we ignore it
	"unable to open database file",    // https://wallarm.atlassian.net/browse/NODE-7590
	"ip source list synchronization",  // https://wallarm.atlassian.net/browse/NODE-7590
	"ip list synchronization done with error",
	"WALLARM:ACL: No such file or directory",
	"--max-errors-in-response",
	// wmcp (NODE-7490) flakes whenever the Cloud returns a bare error
	// string ("Check permissions failed", "Limited by subscription", ...)
	// in place of the expected JSON struct on its config-loader endpoints
	// (get_node_mitigations, list_configs, ...). wmcp logs the parse
	// failure at ERRO level. Not our story (NODE-6240 is wd) and not
	// deterministic — flaky Cloud behavior. Owner: Cloud team / wmcp.
	//
	// wmcp also has a non-atomic-stdout-write race that occasionally
	// interleaves two concurrent log lines byte-for-byte, breaking the
	// JSON envelope. The pattern matches wmcp's message structure
	// ('Error loading cloud config:' + 'failed to parse response')
	// rather than field ordering — both prefixes survive the interleave
	// in the merged buffer and `.*` glides over the garbage between them.
	`Error loading cloud config:.*failed to parse response`,
	// wd's init pipeline can race with the WALLARM_LABELS env propagation:
	// the first register attempt in a fresh container occasionally fires
	// before the env-from-yaml sync, so the label is missing and Cloud
	// returns 'deployCloudNode: label "group" is required'. wd then
	// enters fallback mode and the background pipeline retries register;
	// when WALLARM_LABELS is finally in place the second attempt succeeds
	// and 'node registration done' lands in the logs (which is what the
	// positive AAS api_token case asserts on). The transient first-attempt
	// error is harmless and is the literal expected outcome of the
	// dedicated 'api_token no group' negative case; allowing it here
	// stops the positive case from flapping. Pattern is narrow enough to
	// keep other deployCloudNode errors visible.
	`label \\"group\\" is required for this registration type`,
	// NODE-7601: nsf_download is a best-effort init step (optional: true).
	// All of its failure paths log at warn level and return nil from Start
	// so the legacy register/datasync/ipfeed/apispec flow continues. The
	// warn lines carry a JSON "error" field which the (?i)error regex below
	// would otherwise flag. Allow all nsf_download warn output across
	// subscriptions: the component is meta and never fails the init.
	`"component":"nsf_download".*"message":"manifest fetch failed"`,
	`"component":"nsf_download".*"message":"manifest decrypt failed"`,
	`"component":"nsf_download".*"message":"resource fetch/decrypt failed; skipping"`,
	`"component":"nsf_download".*"message":"no storage configured for region`,
	`"component":"nsf_download".*"message":"derive key failed"`,
	`"component":"nsf_download".*"message":"streaming encrypter init failed"`,
	`"component":"nsf_download".*"message":"license PEM write failed"`,
	`"component":"nsf_download".*"message":"envelope resource skipped: no license PEM in manifest"`,
}

// negativeAllowedErrors is the allow-list used by every case in
// register_negative.go. It extends commonAllowedErrors with markers that
// only legitimately appear when the Cloud rejects the node — e.g. an
// invalid/wrong-host token causes the register job to fail, wd's init
// pipeline records "init stage N failed", drops into fallback mode, and
// every retry of the failing stage re-emits the same line; wstore in turn
// logs that its metrics-exporter refresh hit access-denied (visible side
// of the wstore-stays-alive-in-fallback fix from meganode). On positive
// cases register succeeds first try and these markers do not appear, so
// they MUST stay out of commonAllowedErrors — putting them there would
// silently swallow real Cloud-side regressions that show up only on a
// healthy install.
var negativeAllowedErrors = func() []string {
	out := append([]string{}, commonAllowedErrors...)
	out = append(out,
		// register's own logging when Cloud rejects deployCloudNode
		// (NFR-wrong-host, token-from-other-host).
		"deployCloudNode: not found",
		// wstore at startup couldn't refresh metrics-exporter creds against
		// Cloud (any negative scenario where token isn't fully valid).
		"initial metrics exporter refresh failed",
		// wd's init pipeline aggregates per-stage failure into a single warn,
		// then wd enters fallback so the background pipeline keeps retrying.
		"init stage [0-9]+ failed",
		"init pipeline failed; entering fallback mode",
		"starting in fallback mode",
	)
	return out
}()

// filterUnexpectedErrors returns the subset of lines that do not match any of allowedPatterns
// (each pattern is treated as a regex).
func filterUnexpectedErrors(lines, allowedPatterns []string) []string {
	var unexpected []string
	for _, line := range lines {
		allowed := false
		for _, pattern := range allowedPatterns {
			if matched, _ := regexp.MatchString(pattern, line); matched {
				allowed = true
				break
			}
		}
		if !allowed {
			unexpected = append(unexpected, line)
		}
	}
	return unexpected
}

func (testSuite *RegisterSuite) TestRegisterNode(t provider.T) {
	t.Title("Register Node")
	t.Feature("Register Node")
	t.AllureID("19579")

	var testCases map[string]shared.RegisterNodeCases

	// Allowed error patterns specific to negative cases where registration fails
	// but the container itself starts (init pipeline falls back, side services retry).
	negativeAllowedErrors := append([]string{}, commonAllowedErrors...)
	negativeAllowedErrors = append(negativeAllowedErrors,
		"deployCloudNode: not found",
		"init stage [0-9]+ failed",
		"init pipeline failed; entering fallback mode",
		"starting in fallback mode",
		"initial metrics exporter refresh failed",
		"ip list synchronization done with error", // transient on shutdown / context canceled
		//`job "iplist" failed: can't sync ip list`, // iplist sync transient failures
	)

	testCases = make(map[string]shared.RegisterNodeCases)
	for subscription, tokenTypes := range testSuite.tokens {
		for tokenType, token := range tokenTypes {
			caseName := "Register Node with " + subscription + " " + tokenType
			expectedResult := "\"node registration done\""
			allowedErrors := commonAllowedErrors
			if subscription == "EXPIRED" {
				expectedResult = "subscription limited"
				// EXPIRED subscription deliberately can't run datasync, iplist,
				// ipfeed, apispec, envexp etc — Cloud rejects with either
				// "subscription limited" or per-feature "access denied". wd's
				// init pipeline aggregates the per-job failures into a single
				// fallback warning whose body is reprinted line-by-line by
				// ctlclient.printFallbackWarning to stderr (used by setup.sh's
				// `wd ctl --wait-init`). Allow the per-job error breakdown +
				// the fallback orchestration messages so the EXPIRED case
				// only asserts on the registration-side "subscription limited"
				// substring it actually cares about.
				allowedErrors = append(allowedErrors,
					"subscription limited",
					"init stage [0-9]+ failed",
					"init pipeline failed; entering fallback mode",
					"starting in fallback mode",
					`job "[a-z_]+" failed:.*access denied`,
					`job "[a-z_]+" failed:.*context canceled`,
				)
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
	_ = negativeAllowedErrors // kept for future per-case override; see appendNegativeCases

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
				lastErr     error
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
				// Always fetch logs for diagnostics and expected-result checks
				out, err = testSuite.dockerClient.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true, ShowStderr: true})
				if err != nil {
					stepCtx.Logf("Failed to get container logs: %v", err)
					return false, nil
				}
				defer out.Close()

				var buf bytes.Buffer
				if _, err = stdcopy.StdCopy(&buf, &buf, out); err != nil {
					stepCtx.Logf("Failed to read container logs: %v", err)
					return false, nil
				}
				logs = buf.Bytes()
				strLogs = buf.String()

				// If we expect fail, we check if the expected message is in the logs
				if paramsTest.ExpectFail {
					if strings.Contains(strLogs, paramsTest.ExpectedResult) {
						return true, nil
					}
					return false, fmt.Errorf("expected failure message '%v' not found in logs", paramsTest.ExpectedResult)
				}

				// For success: probe /ready via `docker exec` running a bash
				// /dev/tcp open INSIDE the target container. This bypasses
				// inter-container networking — required for the DinD runner
				// setup where the test container and the node container are
				// on different daemon's bridges (172.17.0.x means different
				// things in each namespace).
				ready, body, probeErr := readyCheckViaExec(ctx, testSuite.dockerClient, resp.ID)
				if probeErr != nil {
					lastErr = fmt.Errorf("ready check: %w (body=%q)", probeErr, body)
					return false, nil
				}
				if !ready {
					lastErr = fmt.Errorf("ready check: not ready yet (body=%q)", body)
					return false, nil
				}
				return true, nil
			})

			stepCtx.WithNewAttachment("Container logs.txt", "text/plain", logs)
			msg := "Error waiting node to be in the expected condition"
			if err != nil {
				msg += fmt.Sprintf(" (poll err: %v)", err)
			}
			if !success && lastErr != nil {
				msg += fmt.Sprintf(" (last ready-check err: %v)", lastErr)
				// On failure, dump the container's full network info so
				// we can see exactly which IP/network containerIP() picked
				// and whether the container has multiple networks.
				if inspect, inspectErr := testSuite.dockerClient.ContainerInspect(context.Background(), resp.ID); inspectErr == nil {
					networkSummary := fmt.Sprintf("legacy_IP=%q networks={", inspect.NetworkSettings.IPAddress)
					for name, n := range inspect.NetworkSettings.Networks {
						networkSummary += fmt.Sprintf(" %s:{IPAddress=%q Gateway=%q NetworkID=%q}", name, n.IPAddress, n.Gateway, n.NetworkID)
					}
					networkSummary += " }"
					msg += " network=" + networkSummary
					stepCtx.Logf("Network dump: %s", networkSummary)
				}
			}
			stepCtx.Require().True(success, msg)
			stepCtx.Require().Contains(strLogs, paramsTest.ExpectedResult, "Error getting expected message: '%v' in log", paramsTest.ExpectedResult)

			// Collect all error-bearing lines, then filter out allowed patterns.
			re = regexp.MustCompile(`(?i)error`)
			var rawErrorLines []string
			for _, line := range strings.Split(strLogs, "\n") {
				if re.MatchString(line) {
					rawErrorLines = append(rawErrorLines, line)
				}
			}

			allowed := paramsTest.AllowedErrorPatterns
			if paramsTest.ExpectedError != "" {
				allowed = append(allowed, regexp.QuoteMeta(paramsTest.ExpectedError))
			}
			errorLines := filterUnexpectedErrors(rawErrorLines, allowed)

			if len(errorLines) > 0 {
				stepCtx.Logf("Found unexpected error messages in logs:\n%s", strings.Join(errorLines, "\n"))
			}
			stepCtx.Require().Empty(errorLines, "Found unexpected error messages in logs")
		})
	}
}

// unregisterAllowedErrors is layered on top of commonAllowedErrors when the
// unregister test scans logs. It allows shutdown-time error lines that come
// from external Wallarm Cloud packages we don't control directly — the
// in-tree wcli/wd jobs use internal.LogJobError which downgrades to Debug
// on context cancellation, but third-party packages bundled into wcli (the
// api-discovery-client cloudclient is the canonical example) still emit
// at error level when their in-flight HTTP call gets cancelled by SIGTERM.
//
// Patterns are kept narrow to the specific component+cause so a real
// regression in the same package can still surface.
var unregisterAllowedErrors = []string{
	// gl.wallarm.com/wallarm-cloud/apid/api-discovery-client/job/cloudclient
	// logs "can't fetch config" at error level when its periodic /v1/user
	// fetch is cancelled by daemon shutdown. Source repo is owned by the
	// Cloud team — fix at source is tracked separately.
	`"component":"api_discovery".*context canceled`,
}

func (testSuite *RegisterSuite) TestUnRegisterNode(t provider.T) {
	t.Title("UnRegister Node")
	t.Feature("UnRegister Node")
	t.AllureID("19580")

	type unregisterCase struct {
		extraEnv        []string
		expectedMessage string
	}
	cases := map[string]unregisterCase{
		"UnRegister Node enabled": {
			extraEnv:        []string{"WALLARM_NODE_UNREGISTER=true"},
			expectedMessage: "node unregistration done",
		},
		"UnRegister Node disabled (default)": {
			extraEnv:        nil,
			expectedMessage: "node unregistration skipped (disabled)",
		},
	}

	for caseName, params := range cases {
		paramsTest := params
		t.WithNewAsyncStep(caseName, func(stepCtx provider.StepCtx) {
			var (
				err         error
				resp        container.CreateResponse
				envVars     []string
				out         io.ReadCloser
				logs        []byte
				strLogs     string
				success     bool
				exec        types.IDResponse
				execInspect container.ExecInspect
				wcliLog     []byte
				ReqInterval = 2 * time.Second
				ReqTimeout  = 120 * time.Second
				execLog     types.HijackedResponse
				re          *regexp.Regexp
				errors      []string
			)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			envVars = []string{
				"WALLARM_API_TOKEN=" + testSuite.tokens["NFR"]["node_token"],
				"WALLARM_API_HOST=audit.api.wallarm.com",
			}
			envVars = append(envVars, paramsTest.extraEnv...)

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

			// Wait for node to be ready (probe inside container — see comment
			// on the helper for why external HTTP doesn't work in DinD).
			success, err = shared.Poll(ReqInterval, ReqTimeout, func() (bool, error) {
				out, err = testSuite.dockerClient.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true, ShowStderr: true})
				if err != nil {
					stepCtx.Logf("Failed to get container logs: %v", err)
					return false, nil
				}
				defer out.Close()

				var buf bytes.Buffer
				if _, err = stdcopy.StdCopy(&buf, &buf, out); err != nil {
					stepCtx.Logf("Failed to read container logs: %v", err)
					return false, nil
				}
				logs = buf.Bytes()

				ready, _, _ := readyCheckViaExec(ctx, testSuite.dockerClient, resp.ID)
				return ready, nil
			})

			stepCtx.WithNewAttachment("Container logs.txt", "text/plain", logs)
			stepCtx.Require().True(success, "Error waiting node to be in the expected condition")

			exec, err = testSuite.dockerClient.ContainerExecCreate(ctx, resp.ID, container.ExecOptions{
				Cmd:          []string{"bash", "-c", "kill $(pgrep wd)"},
				AttachStderr: true,
				AttachStdout: true,
			})
			stepCtx.Require().NoError(err, "Error creating exec")

			execLog, err = testSuite.dockerClient.ContainerExecAttach(ctx, exec.ID, container.ExecStartOptions{})
			stepCtx.Require().NoError(err, "Error starting exec")

			wcliLog, err = io.ReadAll(execLog.Reader)
			if execLog.Reader != nil {
				execLog.Close()
			}
			stepCtx.WithNewAttachment("kill_wcli.txt", "text/plain", wcliLog)

			execInspect, err = testSuite.dockerClient.ContainerExecInspect(ctx, exec.ID)
			stepCtx.Require().NoError(err, "Error executing kill to wcli")
			stepCtx.Require().Equal(execInspect.ExitCode, 0, "Error executing exec")

			success, err = shared.Poll(ReqInterval, ReqTimeout, func() (bool, error) {
				out, err = testSuite.dockerClient.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true, ShowStderr: true})
				if err != nil {
					stepCtx.Logf("Failed to get container logs: %v", err)
					return false, nil
				}
				defer out.Close()

				logs, err = io.ReadAll(out)
				strLogs = string(logs[:])

				if err != nil {
					stepCtx.Logf("Failed to read container logs: %v", err)
					return false, nil
				}
				if strings.Contains(strLogs, paramsTest.expectedMessage) {
					return true, nil
				}
				return false, nil
			})

			stepCtx.WithNewAttachment("wcli.txt", "text/plain", logs)
			stepCtx.Require().True(success, "Error getting expected message: '%v' in log", paramsTest.expectedMessage)

			// Check if there are any errors in the logs, ignoring known allowed patterns.
			// Allow the unregister-specific shutdown patterns on top of the common ones.
			re = regexp.MustCompile(`level\":\"error.*\n`)
			errors = re.FindAllString(strLogs, -1)
			allowed := append([]string{}, commonAllowedErrors...)
			allowed = append(allowed, unregisterAllowedErrors...)
			unexpectedErrors := filterUnexpectedErrors(errors, allowed)
			stepCtx.Require().Empty(unexpectedErrors, "There are errors in the logs. See attachment for details")
		})
	}
}
