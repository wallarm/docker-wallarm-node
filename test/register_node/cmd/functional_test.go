//go:build functional

package cmd

import (
	"gl.wallarm.com/wallarm-node/aio-docker/test/register_node/suites/functional"
	"testing"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
)

type FunctionalSuite struct {
	suite.Suite
}

func (testSuite *FunctionalSuite) TestRegisterNode(t provider.T) {
	t.Parallel()
	testSuite.RunSuite(t, new(functional.RegisterSuite))
}

func TestAllFunctional(t *testing.T) {
	suite.RunSuite(t, new(FunctionalSuite))
}
