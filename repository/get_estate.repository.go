package repository

import (
	"context"

	"github.com/google/uuid"

	"spgo/util"
)

func (r *Repository) GetEstate(ctx context.Context, id uuid.UUID) (EstateEntity, error) {
	var estate EstateEntity

	tx := util.GetTxFromContext(ctx, r.Db)

	if err := tx.WithContext(ctx).Where("id = ?", id).First(&estate).Error; err != nil {
		return EstateEntity{}, err
	}
	return estate, nil
}
