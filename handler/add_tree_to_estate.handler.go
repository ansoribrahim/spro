package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"spgo/generated"
)

func (s *Server) AddTreeToEstate(ctx echo.Context, id openapi_types.UUID) error {
	var req generated.TreeRequest
	var resp generated.TreeResponse

	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	if err := ctx.Validate(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	resp, httpStatus, err := s.Service.AddTreeToEstate(ctx, req, id)
	if err != nil {
		return ctx.JSON(httpStatus, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusCreated, resp)
}
