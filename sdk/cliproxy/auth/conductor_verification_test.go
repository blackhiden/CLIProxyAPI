package auth

import (
	"context"
	"testing"
	"time"

	"github.com/MoeForge/moego/ojson"
	"github.com/blackhiden/CLIProxyAPI/v8/internal/registry"
)

func TestVerificationMaxOutputTokens(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		model    *registry.ModelInfo
		want     int
	}{
		{
			name:     "Antigravity Opus thinking minimum",
			provider: "antigravity",
			model:    registry.LookupModelInfo("claude-opus-4-6-thinking", "antigravity"),
			want:     1040,
		},
		{
			name:     "Antigravity Sonnet thinking minimum",
			provider: "antigravity",
			model:    registry.LookupModelInfo("claude-sonnet-4-6", "antigravity"),
			want:     1040,
		},
		{
			name:     "Codex retains its protocol minimum",
			provider: "codex",
			model:    &registry.ModelInfo{ID: "gpt-5.4", MaxCompletionTokens: 128000},
			want:     16,
		},
		{
			name:     "unknown model uses generic minimum",
			provider: "antigravity",
			want:     16,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.model == nil && tt.name != "unknown model uses generic minimum" {
				t.Fatal("expected model metadata in static registry")
			}
			if got := verificationMaxOutputTokens(tt.provider, tt.model); got != tt.want {
				t.Fatalf("verificationMaxOutputTokens(%q, %#v) = %d, want %d", tt.provider, tt.model, got, tt.want)
			}
		})
	}
}

func TestManager_VerifyInferenceUsesAntigravityThinkingBudget(t *testing.T) {
	ctx := WithSkipPersist(context.Background())
	for _, model := range []string{"claude-opus-4-6-thinking", "claude-sonnet-4-6"} {
		t.Run(model, func(t *testing.T) {
			executor := &unauthorizedRefreshExecutor{id: "antigravity"}
			manager := NewManager(nil, nil, nil)
			manager.RegisterExecutor(executor)
			auth := &Auth{
				ID:       "antigravity-verify-" + model,
				FileName: "antigravity-verify-" + model,
				Provider: "antigravity",
				Status:   StatusActive,
				Metadata: map[string]any{
					"type":         "antigravity",
					"access_token": "test-access-token",
					"project_id":   "test-project",
					"expired":      time.Now().Add(time.Hour).Format(time.RFC3339),
				},
			}
			if _, errRegister := manager.Register(ctx, auth); errRegister != nil {
				t.Fatalf("Register() error = %v", errRegister)
			}
			if _, errVerify := manager.VerifyInference(ctx, auth.ID, model); errVerify != nil {
				t.Fatalf("VerifyInference() error = %v", errVerify)
			}

			executor.mu.Lock()
			payload := append([]byte(nil), executor.lastStreamReq.Payload...)
			executor.mu.Unlock()
			if got := ojson.GetBytesNoCopy(payload, "max_output_tokens").Int(); got != 1040 {
				t.Fatalf("verification max_output_tokens = %d, want 1040; payload=%s", got, payload)
			}
		})
	}
}
