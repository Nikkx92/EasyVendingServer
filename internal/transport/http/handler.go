package http

import "test/internal/services"

type Handler struct {
	auth *services.AuthService
}

func NewHTTPHandler(auth *services.AuthService) *Handler {
	return &Handler{auth: auth}
}
