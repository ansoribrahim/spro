package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"spgo/generated"
)

func (s *Server) GetEstateIdStats(ctx echo.Context, id openapi_types.UUID) error {
	var resp generated.EstateStatsResponse

	resp, err := s.Service.GetEstateStats(ctx.Request().Context(), id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusOK, resp)
}
