package http

import "github.com/gin-gonic/gin"

func RegisterRouters(r *gin.Engine, h *Handler) {
	r.POST("/auth", h.Auth)
}
