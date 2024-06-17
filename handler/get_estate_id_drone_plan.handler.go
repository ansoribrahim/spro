package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"spgo/generated"
)

func (s *Server) GetEstateIdDronePlan(ctx echo.Context, id openapi_types.UUID, params generated.GetEstateIdDronePlanParams) error {

	resp, err := s.Service.GetEstateDronePlan(ctx.Request().Context(), id, params.MaxDistance)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusOK, resp)
}
