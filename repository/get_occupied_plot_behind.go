package repository

import (
	"context"

	"github.com/google/uuid"

	"spgo/util"
)

func (r *Repository) GetOccupiedPlotBehind(ctx context.Context, estateId uuid.UUID, currentOrderNumber int) (*PlotEntity, error) {
	var plot *PlotEntity

	tx := util.GetTxFromContext(ctx, r.Db)

	err := tx.WithContext(ctx).
		Where("estate_id = ? and order_number < ?", estateId, currentOrderNumber).
		Order("order_number desc").
		First(&plot).Error

	if err != nil {
		return nil, err
	}
	return plot, nil
}
