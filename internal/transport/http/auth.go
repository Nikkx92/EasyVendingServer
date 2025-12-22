package http

import (
	"github.com/gin-gonic/gin"
	"time"
)

func (h *Handler) Auth(c *gin.Context) {
	time.Sleep(7 * time.Second)
}
