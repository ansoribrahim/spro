package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"spgo/generated"
)

func (s *Server) PostEstate(ctx echo.Context) error {
	var req generated.EstateRequest
	var resp generated.EstateResponse

	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	if err := ctx.Validate(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	resp, err := s.Service.PostEstate(ctx.Request().Context(), req)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusCreated, resp)
}
