package handler

import (
	"net/http"

	"github.com/gabrielmelo/tg-forward/internal/api/serializer"
	"github.com/gabrielmelo/tg-forward/internal/api/types"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Handle(w http.ResponseWriter, r *http.Request) (*types.DataResponse, *serializer.Error) {
	return &types.DataResponse{Data: types.HealthResponse{Status: "ok"}}, nil
}
