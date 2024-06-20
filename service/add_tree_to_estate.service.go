package service

import (
	"context"
	"errors"
	"math"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"spgo/generated"
	"spgo/repository"
	"spgo/util"
)

func (s *Service) AddTreeToEstate(ctx echo.Context, req generated.TreeRequest, estateId uuid.UUID) (generated.TreeResponse, int, error) {
	resp := generated.TreeResponse{}
	var err error
	tx := s.Db.WithContext(ctx.Request().Context()).Begin()

	nCtx := util.NewTxContext(ctx.Request().Context(), tx)
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
		}
		util.HandleTransaction(tx, err)
	}()

	plot, estate, status, err := s.constructPlot(nCtx, req, estateId)
	if err != nil {
		return generated.TreeResponse{}, status, err
	}

	resp.Id, err = s.Repository.PostPlot(nCtx, *plot)
	if err != nil {
		return generated.TreeResponse{}, http.StatusInternalServerError, err
	}

	opf, err := s.Repository.GetOccupiedPlotForward(nCtx, estate.ID, plot.OrderNumber)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return generated.TreeResponse{}, http.StatusInternalServerError, err
	}

	if opf != nil {
		additionalDistanceGap := 0
		distanceBetweenPlots := opf.OrderNumber - plot.OrderNumber
		if distanceBetweenPlots == 1 {
			treeHeightDifferent := math.Abs(float64(opf.TreeHeight - plot.TreeHeight))
			newDistanceOpf := plot.Distance + int(treeHeightDifferent) + 10
			additionalDistanceGap = newDistanceOpf - opf.Distance
			opf.Distance = newDistanceOpf
		} else {
			remainingDistanceBetweenPlot := (distanceBetweenPlots - 1) * 10
			distanceToOpf := plot.Distance + plot.TreeHeight + 1 + remainingDistanceBetweenPlot
			newDistanceOpf := distanceToOpf + opf.TreeHeight + 1 + 10
			additionalDistanceGap = newDistanceOpf - opf.Distance
			opf.Distance = newDistanceOpf
		}

		_, err = s.Repository.SavePlot(nCtx, *opf)
		if err != nil {
			return generated.TreeResponse{}, http.StatusInternalServerError, err
		}

		err = s.Repository.AdjustPlotForwardDistance(nCtx, estate.ID, opf.OrderNumber, additionalDistanceGap)
		if err != nil {
			return generated.TreeResponse{}, http.StatusInternalServerError, err
		}
	}

	medianTreeHeight, err := s.Repository.GetMedianTreeHeight(nCtx, estate.ID)
	if err != nil {
		return generated.TreeResponse{}, http.StatusInternalServerError, err
	}

	if estate.TreeMinHeight <= 0 {
		estate.TreeMinHeight = plot.TreeHeight
	}

	plotPrev, err := s.Repository.GetPlotByOrderNumber(nCtx, estate.ID, plot.OrderNumber-1)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return generated.TreeResponse{}, http.StatusInternalServerError, err
	}

	plotNext, err := s.Repository.GetPlotByOrderNumber(nCtx, estate.ID, plot.OrderNumber+1)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return generated.TreeResponse{}, http.StatusInternalServerError, err
	}

	if plotPrev != nil {
		estate.TotalDistance += int(math.Abs(float64(plotPrev.TreeHeight - plot.TreeHeight)))
	} else {
		estate.TotalDistance += plot.TreeHeight
	}

	if plotNext != nil {
		estate.TotalDistance += int(math.Abs(float64(plotNext.TreeHeight - plot.TreeHeight)))
	} else {
		estate.TotalDistance += plot.TreeHeight
	}

	estate.TreeCount++
	estate.TreeMaxHeight = int(math.Max(float64(estate.TreeMaxHeight), float64(plot.TreeHeight)))
	estate.TreeMinHeight = int(math.Min(float64(estate.TreeMinHeight), float64(plot.TreeHeight)))
	estate.TreeMedianHeight = medianTreeHeight

	_, err = s.Repository.SaveEstate(nCtx, *estate)
	if err != nil {
		return generated.TreeResponse{}, http.StatusInternalServerError, err
	}

	return resp, http.StatusOK, nil
}

func (s *Service) constructPlot(nCtx context.Context, req generated.TreeRequest, estateId uuid.UUID) (*repository.PlotEntity, *repository.EstateEntity, int, error) {
	_, err := s.Repository.GetPlotByXAndY(nCtx, estateId, req.X, req.Y)
	if err == nil {
		return nil, nil, http.StatusBadRequest, errors.New("plot with coordinate x and y is already occupied")
	}

	estate, err := s.Repository.GetEstate(nCtx, estateId)
	if err != nil {
		return nil, nil, http.StatusNotFound, err
	}

	if req.X > estate.Length || req.Y > estate.Width {
		return nil, nil, http.StatusBadRequest, errors.New("x or y is out of range")
	}

	plot := repository.PlotEntity{
		EstateId:   estate.ID,
		X:          uint16(req.X),
		Y:          uint16(req.Y),
		TreeHeight: req.Height,
		Distance:   0,
	}

	if req.Y%2 == 1 { // Y is odd
		plot.OrderNumber = (req.Y-1)*estate.Length + req.X
	} else { // Y is even
		plot.OrderNumber = (req.Y-1)*estate.Length + (estate.Length - req.X + 1)
	}

	opb, err1 := s.Repository.GetOccupiedPlotBehind(nCtx, estate.ID, plot.OrderNumber)
	if err1 != nil {
		if errors.Is(err1, gorm.ErrRecordNotFound) {
			/*
				if there is no plot behind, then the distance is 10 * order number, which 10 is the width of every plot.
				then to cover tree high and width of this plot we use 10 + tree height
			*/
			distanceToCurrentPlot := (plot.OrderNumber - 1) * 10
			distanceToCoverTree := plot.TreeHeight + 1 + 10
			plot.Distance = distanceToCurrentPlot + distanceToCoverTree
		} else {
			return nil, nil, http.StatusInternalServerError, err1
		}
	} else {
		/*
			if there is other tree, we need to recognise only the previous tree, the nearer one.
			the rest/other tree before previous tree not counted because we can depend on previous tree's distance,
			it already covered all distance.
		*/
		distanceBetweenPlots := plot.OrderNumber - opb.OrderNumber
		if distanceBetweenPlots == 1 {
			/*
				if tree is 1 distance away from previous tree, then we need to cover only the tree height
				difference between previous tree and current tree. because drone doesn't land to the ground.
			*/
			treeHeightDifferent := opb.TreeHeight - plot.TreeHeight
			distanceToCoverTree := int(math.Abs(float64(treeHeightDifferent)) + 10)
			plot.Distance = opb.Distance + distanceToCoverTree
		} else {
			/*
				if tree is more than 1 distance away from previous tree, then we need to cover the distance between
				previous tree and current tree. the distance between previous tree and current tree
				is 10 * distanceBetweenPlots, which 10 is the width of every plot. then to cover tree high and width
				of this plot we use 10 + tree height
			*/
			distancePreviousPlotIncludingDroneLanding := opb.Distance + opb.TreeHeight + 1
			remainingDistanceBetweenPlot := (distanceBetweenPlots - 1) * 10
			distanceToCurrentPlot := distancePreviousPlotIncludingDroneLanding + remainingDistanceBetweenPlot
			distanceToCoverTree := plot.TreeHeight + 1 + 10
			plot.Distance = distanceToCurrentPlot + distanceToCoverTree
		}
	}

	return &plot, &estate, http.StatusCreated, nil
}
