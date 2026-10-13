package management

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// authAPIKeyItem represents a single authentication API key entry.
type authAPIKeyItem struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

// GetAuthAPIKeys returns all available authentication API keys.
// This endpoint is protected by management key authentication and returns
// the keys in the format: [{"name": "<name>", "key": "<API key>"}, ...]
func (h *Handler) GetAuthAPIKeys(c *gin.Context) {
	if h == nil || h.cfg == nil {
		c.JSON(http.StatusOK, []authAPIKeyItem{})
		return
	}

	keys := make([]authAPIKeyItem, 0, len(h.cfg.SDKConfig.APIKeys))
	for _, key := range h.cfg.SDKConfig.APIKeys {
		keys = append(keys, authAPIKeyItem{
			Name: key,
			Key:  key,
		})
	}

	c.JSON(http.StatusOK, keys)
}
