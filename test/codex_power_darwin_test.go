//go:build darwin

package test

import (
	"context"
	"os"
	"testing"

	"github.com/blackhiden/CLIProxyAPI/v8/internal/auth/codex"
)

func TestMain(m *testing.M) {
	codex.ReadSIWCSystemState = func(context.Context) ([]byte, error) {
		return []byte("Current System Capabilities are: CPU Graphics Audio Network\nCurrent Power State: 4\n"), nil
	}
	os.Exit(m.Run())
}
