package management

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
)

func TestGetAuthAPIKeys(t *testing.T) {
	tests := []struct {
		name        string
		apiKeys     []string
		apiKeyNames map[string]string
		expected    []authAPIKeyItem
	}{
		{
			name:     "empty keys",
			apiKeys:  []string{},
			expected: []authAPIKeyItem{},
		},
		{
			name:    "single key",
			apiKeys: []string{"test-key-1"},
			expected: []authAPIKeyItem{
				{Name: "test-key-1", Key: "test-key-1"},
			},
		},
		{
			name:        "multiple keys with persisted names",
			apiKeys:     []string{"key-1", "key-2", "key-3"},
			apiKeyNames: map[string]string{"key-1": "Shell", "key-3": "VSCode"},
			expected: []authAPIKeyItem{
				{Name: "Shell", Key: "key-1"},
				{Name: "key-2", Key: "key-2"},
				{Name: "VSCode", Key: "key-3"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				SDKConfig: config.SDKConfig{
					APIKeys:     tt.apiKeys,
					APIKeyNames: tt.apiKeyNames,
				},
			}
			h := &Handler{cfg: cfg}

			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/v0/management/auth-api-keys", nil)

			h.GetAuthAPIKeys(c)

			if w.Code != http.StatusOK {
				t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
			}

			var result []authAPIKeyItem
			if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d items, got %d", len(tt.expected), len(result))
			}

			for i, item := range result {
				if i >= len(tt.expected) {
					break
				}
				if item.Name != tt.expected[i].Name {
					t.Errorf("item[%d].Name: expected %q, got %q", i, tt.expected[i].Name, item.Name)
				}
				if item.Key != tt.expected[i].Key {
					t.Errorf("item[%d].Key: expected %q, got %q", i, tt.expected[i].Key, item.Key)
				}
			}
		})
	}
}

func TestGetAuthAPIKeys_NilHandler(t *testing.T) {
	var h *Handler

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/v0/management/auth-api-keys", nil)

	h.GetAuthAPIKeys(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var result []authAPIKeyItem
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("expected empty array for nil handler, got %d items", len(result))
	}
}

func TestGetAuthAPIKeys_NilConfig(t *testing.T) {
	h := &Handler{cfg: nil}

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/v0/management/auth-api-keys", nil)

	h.GetAuthAPIKeys(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var result []authAPIKeyItem
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("expected empty array for nil config, got %d items", len(result))
	}
}
