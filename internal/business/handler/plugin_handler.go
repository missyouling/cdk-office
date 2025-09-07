package handler

import (
	"net/http"

	"cdk-office/internal/business/service"
	"github.com/gin-gonic/gin"
)

// PluginHandlerInterface defines the interface for plugin handler
type PluginHandlerInterface interface {
	RegisterPlugin(c *gin.Context)
	UnregisterPlugin(c *gin.Context)
	ListPlugins(c *gin.Context)
	GetPlugin(c *gin.Context)
	EnablePlugin(c *gin.Context)
	DisablePlugin(c *gin.Context)
}

// PluginHandler implements the PluginHandlerInterface
type PluginHandler struct {
	pluginService service.PluginServiceInterface
}

// NewPluginHandler creates a new instance of PluginHandler
func NewPluginHandler() *PluginHandler {
	return &PluginHandler{
		pluginService: service.NewPluginService(),
	}
}

// RegisterPluginRequest represents the request for registering a plugin
type RegisterPluginRequest struct {
	TeamID      string `json:"team_id" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Version     string `json:"version"`
	EntryPoint  string `json:"entry_point" binding:"required"`
	Config      string `json:"config"`
}

// RegisterPlugin handles registering a new plugin
func (h *PluginHandler) RegisterPlugin(c *gin.Context) {
	var req RegisterPluginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to register plugin
	if err := h.pluginService.RegisterPlugin(c.Request.Context(), &service.RegisterPluginRequest{
		TeamID:      req.TeamID,
		Name:        req.Name,
		Description: req.Description,
		Version:     req.Version,
		EntryPoint:  req.EntryPoint,
		Config:      req.Config,
	}); err != nil {
		if err.Error() == "plugin name already exists for this team" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "plugin name already exists for this team"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "plugin registered successfully"})
}

// UnregisterPlugin handles unregistering a plugin
func (h *PluginHandler) UnregisterPlugin(c *gin.Context) {
	pluginID := c.Param("id")
	if pluginID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "plugin id is required"})
		return
	}

	// Call service to unregister plugin
	if err := h.pluginService.UnregisterPlugin(c.Request.Context(), pluginID); err != nil {
		if err.Error() == "plugin not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "plugin not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "plugin unregistered successfully"})
}

// ListPlugins handles listing plugins for a team
func (h *PluginHandler) ListPlugins(c *gin.Context) {
	teamID := c.Query("team_id")
	if teamID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "team id is required"})
		return
	}

	// Call service to list plugins
	plugins, err := h.pluginService.ListPlugins(c.Request.Context(), teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, plugins)
}

// GetPlugin handles retrieving a plugin by ID
func (h *PluginHandler) GetPlugin(c *gin.Context) {
	pluginID := c.Param("id")
	if pluginID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "plugin id is required"})
		return
	}

	// Call service to get plugin
	plugin, err := h.pluginService.GetPlugin(c.Request.Context(), pluginID)
	if err != nil {
		if err.Error() == "plugin not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "plugin not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, plugin)
}

// EnablePlugin handles enabling a plugin
func (h *PluginHandler) EnablePlugin(c *gin.Context) {
	pluginID := c.Param("id")
	if pluginID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "plugin id is required"})
		return
	}

	// Call service to enable plugin
	if err := h.pluginService.EnablePlugin(c.Request.Context(), pluginID); err != nil {
		if err.Error() == "plugin not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "plugin not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "plugin enabled successfully"})
}

// DisablePlugin handles disabling a plugin
func (h *PluginHandler) DisablePlugin(c *gin.Context) {
	pluginID := c.Param("id")
	if pluginID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "plugin id is required"})
		return
	}

	// Call service to disable plugin
	if err := h.pluginService.DisablePlugin(c.Request.Context(), pluginID); err != nil {
		if err.Error() == "plugin not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "plugin not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "plugin disabled successfully"})
}