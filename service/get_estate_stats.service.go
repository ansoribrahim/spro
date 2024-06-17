package service

import (
	"context"

	"github.com/google/uuid"

	"spgo/generated"
)

func (s *Service) GetEstateStats(ctx context.Context, id uuid.UUID) (generated.EstateStatsResponse, error) {
	estate, err := s.Repository.GetEstate(ctx, id)
	if err != nil {
		return generated.EstateStatsResponse{}, err
	}

	return generated.EstateStatsResponse{
		Max:    &estate.TreeMaxHeight,
		Min:    &estate.TreeMinHeight,
		Median: &estate.TreeMedianHeight,
		Count:  &estate.TreeCount,
	}, nil
}
