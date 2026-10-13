// FORK-ONLY: This test ensures the CPA fork's persistent Live Flow feature
// is not accidentally removed by upstream merges.
// See docs/MERGE-PRESERVATION-fork-fixes.md for merge policy.
package api

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestLiveFlowFeaturePersistence verifies that all critical Live Flow files
// and markers exist. This test will fail if an upstream merge accidentally
// removes the fork-only Live Flow feature.
func TestLiveFlowFeaturePersistence(t *testing.T) {
	// Required files that must exist
	requiredFiles := []string{
		"internal/api/flowviz.go",
		"internal/api/flowviz_test.go",
		"internal/api/flowviz_bench_test.go",
		"internal/api/flowviz_conversation_recovery_test.go",
	}

	projectRoot := findProjectRoot(t)

	for _, file := range requiredFiles {
		path := filepath.Join(projectRoot, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("FORK-ONLY Live Flow file missing: %s\nThis file is required for the fork's persistent Live Flow feature.\nSee docs/MERGE-PRESERVATION-fork-fixes.md for restoration instructions.", file)
		}
	}

	// Check for FORK-ONLY markers in flowviz.go
	flowvizPath := filepath.Join(projectRoot, "internal/api/flowviz.go")
	if data, err := os.ReadFile(flowvizPath); err == nil {
		content := string(data)
		if !strings.Contains(content, "FORK-ONLY") {
			t.Error("flowviz.go is missing the FORK-ONLY marker at the top.\nThis marker identifies the file as part of the fork's persistent feature.\nRestore the marker: '// FORK-ONLY: This file implements the CPA fork's persistent Live Flow feature.'")
		}
	}

	// Check for config field in config.go
	configPath := filepath.Join(projectRoot, "internal/config/config.go")
	if data, err := os.ReadFile(configPath); err == nil {
		content := string(data)
		if !strings.Contains(content, "CPAForkLiveFlowEnabled") {
			t.Error("config.go is missing the CPAForkLiveFlowEnabled field.\nThis field is required for the fork's Live Flow feature.\nSee docs/MERGE-PRESERVATION-fork-fixes.md for the required code.")
		}
		if !strings.Contains(content, "cpa-fork-liveflow-enabled") {
			t.Error("config.go is missing the 'cpa-fork-liveflow-enabled' yaml tag.\nThis tag is required for config file parsing.\nSee docs/MERGE-PRESERVATION-fork-fixes.md for the required code.")
		}
	}

	// Check for normalization function
	normPath := filepath.Join(projectRoot, "internal/config/config_normalization.go")
	if data, err := os.ReadFile(normPath); err == nil {
		content := string(data)
		if !strings.Contains(content, "NormalizeLiveFlowConfig") {
			t.Error("config_normalization.go is missing the NormalizeLiveFlowConfig function.\nThis function provides backward compatibility for the legacy config key.\nSee docs/MERGE-PRESERVATION-fork-fixes.md for the required code.")
		}
	}

	// Check for server integration in server.go
	serverPath := filepath.Join(projectRoot, "internal/api/server.go")
	if data, err := os.ReadFile(serverPath); err == nil {
		content := string(data)
		if !strings.Contains(content, "CPAForkLiveFlowEnabled") {
			t.Error("server.go is not checking cfg.CPAForkLiveFlowEnabled.\nThe server must check the fork-specific config key, not the legacy key.\nSee docs/MERGE-PRESERVATION-fork-fixes.md for the required code.")
		}
		if !strings.Contains(content, "newFlowHub()") {
			t.Error("server.go is not creating the flow hub.\nThe Live Flow feature requires hub instantiation when enabled.\nSee docs/MERGE-PRESERVATION-fork-fixes.md for the required code.")
		}
	}
}

// findProjectRoot walks up the directory tree to find the project root
// (the directory containing go.mod).
func findProjectRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find project root (go.mod not found)")
		}
		dir = parent
	}
}
