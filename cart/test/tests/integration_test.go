package tests

import (
	"testing"

	"github.com/stretchr/testify/suite"

	t_suite "route256/cart/test/suite"
)

func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(t_suite.TSuite))
}
