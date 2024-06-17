package repository

import (
	"context"

	"github.com/google/uuid"

	"spgo/util"
)

func (r *Repository) GetPlotByXAndY(ctx context.Context, estateId uuid.UUID, x int, y int) (*uuid.UUID, error) {
	var plot PlotEntity

	tx := util.GetTxFromContext(ctx, r.Db)

	if err := tx.WithContext(ctx).Where("estate_id = ? and x = ? and y = ?", estateId, x, y).First(&plot).Error; err != nil {
		return nil, err
	}
	return &plot.ID, nil
}
