//go:build functional

package functional

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/ozontech/allure-go/pkg/framework/provider"

	"gl.wallarm.com/wallarm-node/aio-docker/test/register_node/shared"
)

// pickNSFToken returns the QA-curated NSF token from W_TEST_TOKENS.
// Prefers node_token (the dedicated subscription's primary token type),
// falls back to api_token only if node_token isn't present in the
// bundle. Returns ok=false when the "NSF" subscription is missing —
// downstream tests then t.Skip, because:
//
//   - The "every token registers, old init path works" assertion is
//     covered by TestRegisterNode's matrix + commonAllowedErrors
//     (nsf_download warns on 404 for legacy api_token UUIDs and the
//     allow-list silently passes those through).
//   - NSF-specific assertions (manifest decoded, no decrypt drift) only
//     make sense against a token whose manifest is actually provisioned
//     in audit — that's exactly the "NSF" subscription bundle.
func (testSuite *RegisterSuite) pickNSFToken() (tokenType, token string, ok bool) {
	tokens, has := testSuite.tokens["NSF"]
	if !has {
		return "", "", false
	}
	if tok, has := tokens["node_token"]; has {
		return "node_token", tok, true
	}
	for tt, tok := range tokens {
		return tt, tok, true
	}
	return "", "", false
}

// runOnce spins up one container with the supplied env, polls readiness via
// `wd ctl status` (same primitive as TestRegisterNode), drains the logs into
// a string, and tears the container down. Returns combined stdout+stderr
// regardless of readiness — even an un-ready container's logs are useful for
// assertions about which init-step actually ran.
func (testSuite *RegisterSuite) runOnce(stepCtx provider.StepCtx, envVars []string, readyTimeout time.Duration) (logs string, ready bool) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resp, err := testSuite.dockerClient.ContainerCreate(ctx, &container.Config{
		Image: testSuite.imageName,
		Tty:   false,
		Env:   envVars,
	}, nodeHostConfig(), nil, nil, "")
	stepCtx.Require().NoError(err, "container create")
	defer func() {
		_ = testSuite.dockerClient.ContainerRemove(
			context.Background(), resp.ID,
			container.RemoveOptions{RemoveVolumes: true, Force: true})
	}()

	stepCtx.Require().NoError(
		testSuite.dockerClient.ContainerStart(ctx, resp.ID, container.StartOptions{}),
		"container start")

	ready, _ = shared.Poll(2*time.Second, readyTimeout, func() (bool, error) {
		r, _, _ := readyCheckViaExec(ctx, testSuite.dockerClient, resp.ID)
		return r, nil
	})

	out, err := testSuite.dockerClient.ContainerLogs(
		ctx, resp.ID,
		container.LogsOptions{ShowStdout: true, ShowStderr: true})
	if err == nil {
		defer out.Close()
		var buf bytes.Buffer
		if _, copyErr := stdcopy.StdCopy(&buf, &buf, out); copyErr != nil && copyErr != io.EOF {
			stepCtx.Logf("log read: %v", copyErr)
		}
		logs = buf.String()
	}
	stepCtx.WithNewAttachment("container logs", "text/plain", []byte(logs))
	return logs, ready
}

// TestNSFDownloadDisabled verifies the operator opt-out path: when
// WALLARM_NSF_DISABLED=true the nsf_download step exits before any HTTP
// activity (skip log present, manifest-decoded log absent), and the
// legacy register/datasync/ipfeed/apispec pipeline still produces a
// healthy node ("node registration done" + ready).
func (testSuite *RegisterSuite) TestNSFDownloadDisabled(t provider.T) {
	t.Title("nsf_download honours WALLARM_NSF_DISABLED=true")
	t.Feature("Register Node")
	t.AllureID("NODE-7601-disabled")

	tokenType, token, ok := testSuite.pickNSFToken()
	if !ok {
		t.Skip(`"NSF" subscription not present in W_TEST_TOKENS — NSF-specific test cannot run`)
		return
	}

	t.WithNewStep("disable NSF, register with NSF/"+tokenType, func(stepCtx provider.StepCtx) {
		envVars := []string{
			"WALLARM_API_TOKEN=" + token,
			"WALLARM_API_HOST=" + apiHost(),
			"WALLARM_NSF_DISABLED=true",
		}
		if tokenType == "api_token" {
			envVars = append(envVars, "WALLARM_LABELS=group=testWallarm")
		}

		logs, ready := testSuite.runOnce(stepCtx, envVars, 80*time.Second)

		stepCtx.Require().True(ready, "node did not reach ready state with WALLARM_NSF_DISABLED=true")
		stepCtx.Require().Contains(logs, "node registration done",
			"legacy register did not complete; opt-out must not break the normal pipeline")
		stepCtx.Require().Contains(logs, "nsf_download disabled by WALLARM_NSF_DISABLED",
			"expected nsf_download skip log not found")
		stepCtx.Require().NotContains(logs, `"message":"manifest decoded"`,
			"nsf_download fetched a manifest despite WALLARM_NSF_DISABLED=true")
	})
}

// TestNSFDownloadEnabled is the dedicated positive test for nsf_download.
// It runs against the "NSF" subscription token whose manifest is curated
// in the audit deployment, asserts the success path end-to-end, and
// fails if NSF-specific drift appears (manifest didn't decode, or
// per-resource decrypt/md5/unwrap failure).
//
// Coverage of "old tokens still register fine when their NSF manifest
// is absent" is intentionally NOT here — that is what TestRegisterNode
// asserts via commonAllowedErrors (the nsf_download warn lines are
// allowed for legacy api_token/EXPIRED cases).
func (testSuite *RegisterSuite) TestNSFDownloadEnabled(t provider.T) {
	t.Title("nsf_download default path against audit cloud")
	t.Feature("Register Node")
	t.AllureID("NODE-7601-enabled")

	tokenType, token, ok := testSuite.pickNSFToken()
	if !ok {
		t.Skip(`"NSF" subscription not present in W_TEST_TOKENS — NSF-specific test cannot run`)
		return
	}

	t.WithNewStep("default NSF, register with NSF/"+tokenType, func(stepCtx provider.StepCtx) {
		envVars := []string{
			"WALLARM_API_TOKEN=" + token,
			"WALLARM_API_HOST=" + apiHost(),
		}
		if tokenType == "api_token" {
			envVars = append(envVars, "WALLARM_LABELS=group=testWallarm")
		}

		logs, ready := testSuite.runOnce(stepCtx, envVars, 80*time.Second)
		stepCtx.Require().True(ready, "node did not reach ready state")
		stepCtx.Require().Contains(logs, "node registration done",
			"legacy register did not complete")
		stepCtx.Require().Contains(logs, `"message":"manifest decoded"`,
			`"NSF" subscription token is expected to have a provisioned manifest in audit; got none`)

		// nsf_download reported success — must NOT have logged per-resource
		// failures or md5/unwrap errors.
		stepCtx.Require().NotContains(logs, `"message":"resource fetch/decrypt failed; skipping"`,
			"nsf_download decoded manifest but a resource failed unexpectedly")
		stepCtx.Require().NotContains(logs, `"reason":"md5_mismatch"`,
			"nsf_download saw md5 mismatch — fixture/cloud data drift")
		stepCtx.Require().NotContains(logs, `"reason":"unwrap"`,
			"nsf_download license-envelope unwrap failed — RSA / format drift")
	})
}
