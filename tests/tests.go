package tests

import (
	"os"
	"testing"
)

const E2ETestEnvVar = "AZURE_REST_API_INDEX_E2E"

func E2EPrecheck(t *testing.T) {
	if os.Getenv(E2ETestEnvVar) == "" {
		t.Skipf("The E2E toggle %s not set", E2ETestEnvVar)
	}
}
