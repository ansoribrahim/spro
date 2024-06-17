// This file contains the interfaces for the repository layer.
// The repository layer is responsible for interacting with the database.
// For testing purpose we will generate mock implementations of these
// interfaces using mockgen. See the Makefile for more information.
package repository

import (
	"context"

	"github.com/google/uuid"
)

type RepositoryInterface interface {
	PostEstate(ctx context.Context, entity EstateEntity) (*uuid.UUID, error)
	GetEstate(ctx context.Context, id uuid.UUID) (EstateEntity, error)
	PostPlot(ctx context.Context, entity PlotEntity) (*uuid.UUID, error)
	SavePlot(ctx context.Context, entity PlotEntity) (*uuid.UUID, error)
	SaveEstate(ctx context.Context, entity EstateEntity) (*uuid.UUID, error)
	GetPlotByXAndY(ctx context.Context, estateId uuid.UUID, x int, y int) (*uuid.UUID, error)
	GetOccupiedPlotBehind(ctx context.Context, estateId uuid.UUID, currentOrderNumber int) (*PlotEntity, error)
	GetOccupiedPlotForward(ctx context.Context, estateId uuid.UUID, currentOrderNumber int) (*PlotEntity, error)
	AdjustPlotForwardDistance(ctx context.Context, estateId uuid.UUID, currentOrderNumber int, additionalDistanceGap int) error
	GetMedianTreeHeight(ctx context.Context, estateID uuid.UUID) (int, error)
	GetPlotByOrderNumber(ctx context.Context, estateId uuid.UUID, orderNumber int) (*PlotEntity, error)
	GetPlotByDistance(ctx context.Context, estateId uuid.UUID, distance int) (*PlotEntity, error)
}
