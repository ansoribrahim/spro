package service

import (
	"context"

	"spgo/generated"
	"spgo/repository"
)

func (s *Service) PostEstate(ctx context.Context, req generated.EstateRequest) (generated.EstateResponse, error) {
	resp := generated.EstateResponse{}
	var err error

	// Total distance when there is no tree would be 10 each plot
	// totalDistance should be width * length * 10
	estate := repository.EstateEntity{
		Width:         req.Width,
		Length:        req.Length,
		TotalDistance: req.Width * req.Length * 10,
	}

	resp.Id, err = s.Repository.PostEstate(ctx, estate)
	if err != nil {
		return generated.EstateResponse{}, err
	}

	return resp, nil
}
