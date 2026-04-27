//go:build functional && positive_only

package functional

import (
	"gl.wallarm.com/wallarm-node/aio-docker/test/register_node/shared"
)

func appendNegativeCases(
	_ map[string]shared.RegisterNodeCases,
	_ map[string]map[string]string,
	_ []string,
) {
}