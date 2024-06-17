package repository

import (
	"context"

	"github.com/google/uuid"

	"spgo/util"
)

func (r *Repository) GetPlotByOrderNumber(ctx context.Context, estateId uuid.UUID, orderNumber int) (*PlotEntity, error) {
	var plot *PlotEntity

	tx := util.GetTxFromContext(ctx, r.Db)

	if err := tx.WithContext(ctx).Where("estate_id = ? and order_number = ? ", estateId, orderNumber).First(&plot).Error; err != nil {
		return nil, err
	}
	return plot, nil
}
