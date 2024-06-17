package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"spgo/util"
)

func (r *Repository) AdjustPlotForwardDistance(ctx context.Context, estateId uuid.UUID, currentOrderNumber int, additionalDistanceGap int) error {
	tx := util.GetTxFromContext(ctx, r.Db)
	err := tx.WithContext(ctx).
		Model(&PlotEntity{}).
		Where("estate_id = ? AND order_number > ?", estateId, currentOrderNumber).
		UpdateColumn("distance", gorm.Expr("distance + ?", additionalDistanceGap)).
		Error
	if err != nil {
		return err
	}

	return nil
}
