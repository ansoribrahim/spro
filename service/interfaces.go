package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"spgo/generated"
)

type ServiceInterface interface {
	PostEstate(ctx context.Context, req generated.EstateRequest) (generated.EstateResponse, error)
	AddTreeToEstate(ctx echo.Context, req generated.TreeRequest, id uuid.UUID) (generated.TreeResponse, int, error)
	GetEstateStats(ctx context.Context, id uuid.UUID) (generated.EstateStatsResponse, error)
	GetEstateDronePlan(ctx context.Context, id uuid.UUID, maxDistance *int) (generated.DronePlanResponse, error)
}
