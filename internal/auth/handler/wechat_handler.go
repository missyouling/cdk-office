package handler

import (
	"net/http"

	"cdk-office/internal/auth/service"
	"github.com/gin-gonic/gin"
)

// WeChatHandlerInterface defines the interface for WeChat authentication handler
type WeChatHandlerInterface interface {
	WeChatLogin(c *gin.Context)
}

// WeChatHandler implements the WeChatHandlerInterface
type WeChatHandler struct {
	wechatService service.WeChatServiceInterface
}

// NewWeChatHandler creates a new instance of WeChatHandler
func NewWeChatHandler() *WeChatHandler {
	return &WeChatHandler{
		wechatService: service.NewWeChatService(),
	}
}

// WeChatLoginRequest represents the request for WeChat login
type WeChatLoginRequest struct {
	Code string `json:"code" binding:"required"`
}

// WeChatLogin handles WeChat login
func (h *WeChatHandler) WeChatLogin(c *gin.Context) {
	var req WeChatLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to login user via WeChat
	resp, err := h.wechatService.WeChatLogin(c.Request.Context(), req.Code)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}