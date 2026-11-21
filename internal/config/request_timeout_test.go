package config

import (
	"math"
	"testing"
	"time"
)

func TestParseRequestDuration(t *testing.T) {
	tests := []struct {
		raw  string
		want time.Duration
	}{
		{"", 0},
		{"   ", 0},
		{"0", 0},
		{"0s", 0},
		{"none", 0},
		{"UNLIMITED", 0},
		{"disabled", 0},
		{"off", 0},
		{"never", 0},
		{"10m", 10 * time.Minute},
		{"600s", 600 * time.Second},
		{"1500ms", 1500 * time.Millisecond},
		{"600", 600 * time.Second},
		{"-1", 0},
		{"-5s", 0},
		{"garbage", 0},
		{"9223372036854775807", time.Duration(math.MaxInt64)},
	}
	for _, tt := range tests {
		if got := ParseRequestDuration(tt.raw); got != tt.want {
			t.Errorf("ParseRequestDuration(%q) = %v, want %v", tt.raw, got, tt.want)
		}
	}
}

func TestRequestTimeoutDurationHelpers(t *testing.T) {
	var nilCfg *SDKConfig
	if got := nilCfg.RequestTimeoutDuration(); got != 0 {
		t.Fatalf("nil RequestTimeoutDuration = %v, want 0", got)
	}
	if got := nilCfg.HardRequestDeadlineDuration(); got != 0 {
		t.Fatalf("nil HardRequestDeadlineDuration = %v, want 0", got)
	}

	cfg := &SDKConfig{
		RequestTimeout:            "10m",
		HardRequestDeadline:       "30s",
		CodexRequestTimeout:       "2m",
		ClaudeHardRequestDeadline: "45s",
	}
	if got := cfg.RequestTimeoutDuration(); got != 10*time.Minute {
		t.Fatalf("global budget = %v, want 10m", got)
	}
	if got := cfg.RequestTimeoutDurationForProvider("codex"); got != 2*time.Minute {
		t.Fatalf("codex budget = %v, want 2m", got)
	}
	if got := cfg.RequestTimeoutDurationForProvider("claude"); got != 10*time.Minute {
		t.Fatalf("claude budget = %v, want global 10m", got)
	}
	if got := cfg.HardRequestDeadlineDurationForProvider("claude"); got != 45*time.Second {
		t.Fatalf("claude deadline = %v, want 45s", got)
	}
	if got := cfg.HardRequestDeadlineDurationForProvider("codex"); got != 30*time.Second {
		t.Fatalf("codex deadline = %v, want global 30s", got)
	}
}
