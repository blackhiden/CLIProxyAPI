//go:build windows

package claude

import (
	"context"

	baseauth "github.com/blackhiden/CLIProxyAPI/v8/internal/auth"
)

// CheckClaudeWindowsReadiness checks whether Windows is ready for Claude OAuth renewal.
var CheckClaudeWindowsReadiness = func(ctx context.Context) error {
	return baseauth.CheckWindowsRefreshReadiness(ctx, "Claude")
}

func claudeRefreshReadiness(ctx context.Context) error {
	return CheckClaudeWindowsReadiness(ctx)
}
