package handlers

import (
	"testing"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v8/sdk/config"
	log "github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"golang.org/x/net/context"
)

func TestRequestTimeoutKeepAliveResolution(t *testing.T) {
	tests := []struct {
		name          string
		cfg           *config.SDKConfig
		streamWant    time.Duration
		nonStreamWant time.Duration
		preFTWant     bool
	}{
		{"disabled", &config.SDKConfig{}, 0, 0, false},
		{"short budget floor", &config.SDKConfig{RequestTimeout: "2s"}, time.Second, time.Second, true},
		{"long budget cap", &config.SDKConfig{RequestTimeout: "10m"}, 30 * time.Second, 30 * time.Second, true},
		{"medium budget", &config.SDKConfig{RequestTimeout: "90s"}, 30 * time.Second, 30 * time.Second, true},
		{"explicit stream override", &config.SDKConfig{RequestTimeout: "10m", Streaming: config.StreamingConfig{KeepAliveSeconds: 5}}, 5 * time.Second, 30 * time.Second, true},
		{"explicit nonstream override", &config.SDKConfig{RequestTimeout: "10m", NonStreamKeepAliveInterval: 7}, 30 * time.Second, 7 * time.Second, true},
		{"explicit preft without budget", &config.SDKConfig{Streaming: config.StreamingConfig{KeepAliveSeconds: 5, SendKeepAliveBeforeFT: true}}, 5 * time.Second, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StreamingKeepAliveInterval(tt.cfg); got != tt.streamWant {
				t.Errorf("streaming interval = %v, want %v", got, tt.streamWant)
			}
			if got := NonStreamingKeepAliveInterval(tt.cfg); got != tt.nonStreamWant {
				t.Errorf("non-streaming interval = %v, want %v", got, tt.nonStreamWant)
			}
			if got := StreamingSendKeepAliveBeforeFT(tt.cfg); got != tt.preFTWant {
				t.Errorf("pre-FT = %v, want %v", got, tt.preFTWant)
			}
		})
	}
}

func TestLogRequestBudgetExceeded(t *testing.T) {
	hook := logtest.NewGlobal()
	defer hook.Reset()
	originalNow := requestTimeoutNow
	defer func() { requestTimeoutNow = originalNow }()

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	requestTimeoutNow = func() time.Time { return start.Add(11 * time.Minute) }
	cfg := &config.SDKConfig{RequestTimeout: "10m"}

	logRequestBudgetExceeded(cfg, "openai", "model-a", start)
	entries := hook.AllEntries()
	if len(entries) != 1 {
		t.Fatalf("log entries = %d, want 1", len(entries))
	}
	if entries[0].Level != log.WarnLevel {
		t.Fatalf("level = %v, want warn", entries[0].Level)
	}
	for _, key := range []string{"handler_type", "model", "budget", "elapsed"} {
		if _, ok := entries[0].Data[key]; !ok {
			t.Errorf("missing log field %q", key)
		}
	}
	if entries[0].Data["model"] != "model-a" {
		t.Errorf("model field = %v", entries[0].Data["model"])
	}

	hook.Reset()
	logRequestBudgetExceeded(&config.SDKConfig{}, "openai", "model-a", start)
	if entries := hook.AllEntries(); len(entries) != 0 {
		t.Fatalf("disabled budget produced log entries: %v", entries)
	}
}

func TestWithHardRequestDeadline(t *testing.T) {
	h := &BaseAPIHandler{Cfg: &config.SDKConfig{HardRequestDeadline: "1h", CodexHardRequestDeadline: "2m"}}
	ctx, cancel := h.withHardRequestDeadline(context.Background(), []string{"codex"})
	defer cancel()
	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("expected hard deadline on context")
	}
	if remaining := time.Until(deadline); remaining <= 0 || remaining > 2*time.Minute {
		t.Fatalf("deadline remaining = %v, want provider override 2m", remaining)
	}

	disabled := &BaseAPIHandler{Cfg: &config.SDKConfig{}}
	ctx, cancel = disabled.withHardRequestDeadline(context.Background(), []string{"codex"})
	defer cancel()
	if _, ok := ctx.Deadline(); ok {
		t.Fatal("disabled hard deadline must not add a context deadline")
	}
}

func TestWithStreamingBootstrapDeadlineDisabled(t *testing.T) {
	base := context.Background()
	ctx, stop := withStreamingBootstrapDeadline(base, &config.SDKConfig{})
	defer stop()
	if ctx != base {
		t.Fatal("disabled hard deadline must reuse the original context")
	}
}
