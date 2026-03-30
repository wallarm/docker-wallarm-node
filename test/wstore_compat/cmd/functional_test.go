//go:build functional

package cmd

import (
	"testing"

	"gl.wallarm.com/wallarm-node/aio-docker/test/wstore_compat/suites/wstore"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
)

type FunctionalSuite struct {
	suite.Suite
}

func (s *FunctionalSuite) TestWstoreCompat(t provider.T) {
	s.RunSuite(t, new(wstore.WstoreCompatSuite))
}

func TestAllFunctional(t *testing.T) {
	suite.RunSuite(t, new(FunctionalSuite))
}
