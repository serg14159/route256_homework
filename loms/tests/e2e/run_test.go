package suite

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// TestIntegrationSuite
func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(TSuite))
}
