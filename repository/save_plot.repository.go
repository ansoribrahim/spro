package repository

import (
	"context"

	"github.com/google/uuid"

	"spgo/util"
)

func (r *Repository) SavePlot(ctx context.Context, entity PlotEntity) (*uuid.UUID, error) {
	tx := util.GetTxFromContext(ctx, r.Db)
	err := tx.WithContext(ctx).Save(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity.ID, nil
}
