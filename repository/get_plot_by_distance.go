package repository

import (
	"context"

	"github.com/google/uuid"

	"spgo/util"
)

func (r *Repository) GetPlotByDistance(ctx context.Context, estateId uuid.UUID, distance int) (*PlotEntity, error) {
	var plot *PlotEntity

	tx := util.GetTxFromContext(ctx, r.Db)

	if err := tx.WithContext(ctx).
		Where("estate_id = ? and distance <= ? ", estateId, distance).
		First(&plot).Error; err != nil {
		return nil, err
	}
	return plot, nil
}
