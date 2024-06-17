package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"spgo/generated"
)

func (s *Service) GetEstateDronePlan(ctx context.Context, estateId uuid.UUID, maxDistance *int) (generated.DronePlanResponse, error) {
	resp := generated.DronePlanResponse{}
	estate, err := s.Repository.GetEstate(ctx, estateId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return generated.DronePlanResponse{}, errors.New("estate not found")
		}
		return generated.DronePlanResponse{}, err
	}

	resp.Distance = &estate.TotalDistance

	if maxDistance != nil {
		var x, y uint16
		var plotDistance int
		*maxDistance--
		plot, err := s.Repository.GetPlotByDistance(ctx, estateId, *maxDistance)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// If plot not found, set plot distance to 0 (start from the beginning)
				plotDistance = 0
				x = 1 // Start x from 1
				y = 1 // Start y from 1
			} else {
				return generated.DronePlanResponse{}, err
			}
		} else {
			plotDistance = plot.Distance
			x = plot.X
			y = plot.Y
		}

		remainBattery := *maxDistance - plotDistance
		remainPlot := remainBattery / 10

		if int(x)+remainPlot <= estate.Length {
			x = uint16(int(x) + remainPlot)
		} else {
			remainPlot -= (estate.Length - int(x))
			y++
			for remainPlot > estate.Length {
				remainPlot -= estate.Length
				y++
			}
			x = uint16(remainPlot)
		}

		xInt := int(x)
		yInt := int(y)
		resp.Rest = &struct {
			X *int `json:"x,omitempty"`
			Y *int `json:"y,omitempty"`
		}{
			X: &xInt,
			Y: &yInt,
		}
	}

	return resp, nil
}
