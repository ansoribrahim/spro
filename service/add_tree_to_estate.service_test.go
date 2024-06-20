package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"spgo/generated"
	"spgo/repository"
)

func TestService_AddTreeToEstate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUUID := uuid.New()
	mockEstateID := uuid.New()
	mockTime := time.Now()

	tests := []struct {
		name           string
		prepareMocks   func(mockRepo *repository.MockRepositoryInterface, mock sqlmock.Sqlmock)
		isPanicTest    bool
		request        generated.TreeRequest
		expectedResp   generated.TreeResponse
		expectedStatus int
		expectedErr    error
	}{
		{
			name: "Successful AddTreeToEstate",
			prepareMocks: func(mockRepo *repository.MockRepositoryInterface, mock sqlmock.Sqlmock) {
				mockPlot := repository.PlotEntity{
					ID:          uuid.New(),
					EstateId:    mockEstateID,
					X:           1,
					Y:           2,
					TreeHeight:  10,
					OrderNumber: 1,
					Distance:    20,
					CreatedAt:   mockTime,
				}
				mockEstate := repository.EstateEntity{
					ID:               mockEstateID,
					Length:           5,
					Width:            10,
					TotalDistance:    20,
					TreeMinHeight:    5,
					TreeMaxHeight:    15,
					TreeCount:        1,
					TreeMedianHeight: 10,
				}
				mockRepo.EXPECT().GetPlotByXAndY(gomock.Any(), mockEstateID, 1, 2).Return(nil, errors.New("not found"))
				mockRepo.EXPECT().GetEstate(gomock.Any(), mockEstateID).Return(mockEstate, nil)
				mockRepo.EXPECT().PostPlot(gomock.Any(), gomock.Any()).Return(&mockUUID, nil)
				mockRepo.EXPECT().GetOccupiedPlotBehind(gomock.Any(), mockEstateID, 10).Return(nil, gorm.ErrRecordNotFound)
				mockRepo.EXPECT().GetOccupiedPlotForward(gomock.Any(), mockEstateID, 10).Return(nil, gorm.ErrRecordNotFound)
				mockRepo.EXPECT().GetMedianTreeHeight(gomock.Any(), mockEstateID).Return(10, nil)
				mockRepo.EXPECT().SaveEstate(gomock.Any(), gomock.Any()).Return(nil, nil)
				mockRepo.EXPECT().GetPlotByOrderNumber(gomock.Any(), gomock.Any(), gomock.Any()).Return(&repository.PlotEntity{ID: mockPlot.ID}, nil).AnyTimes()

				rows := sqlmock.NewRows([]string{"id"}).AddRow(mockPlot.ID)
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(
					`INSERT INTO "plots" ("estate_id","x","y","distance","order_number","tree_height","created_at") VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING "id"`)).
					WithArgs(mockPlot.EstateId, mockPlot.X, mockPlot.Y, mockPlot.Distance, mockPlot.OrderNumber, mockPlot.TreeHeight, mockPlot.CreatedAt).
					WillReturnRows(rows)
				mock.ExpectCommit()
			},
			request: generated.TreeRequest{
				X:      1,
				Y:      2,
				Height: 10,
			},
			expectedResp: generated.TreeResponse{
				Id: &mockUUID,
			},
			expectedStatus: http.StatusOK,
			expectedErr:    nil,
		},
		{
			name: "Repository Error",
			prepareMocks: func(mockRepo *repository.MockRepositoryInterface, mock sqlmock.Sqlmock) {
				mockRepo.EXPECT().GetPlotByXAndY(gomock.Any(), mockEstateID, 1, 2).Return(nil, errors.New("some error"))
				mockRepo.EXPECT().GetEstate(gomock.Any(), mockEstateID).Return(repository.EstateEntity{}, errors.New("some error"))
				mock.ExpectBegin()
				mock.ExpectRollback()
			},
			request: generated.TreeRequest{
				X:      1,
				Y:      2,
				Height: 10,
			},
			expectedResp:   generated.TreeResponse{},
			expectedStatus: http.StatusNotFound,
			expectedErr:    errors.New("some error"),
		},
		{
			name: "Plot Already Exists",
			prepareMocks: func(mockRepo *repository.MockRepositoryInterface, mock sqlmock.Sqlmock) {
				mockRepo.EXPECT().GetPlotByXAndY(gomock.Any(), mockEstateID, 1, 2).Return(&mockUUID, nil)
			},
			request: generated.TreeRequest{
				X:      1,
				Y:      2,
				Height: 10,
			},
			expectedResp:   generated.TreeResponse{},
			expectedStatus: http.StatusBadRequest,
			expectedErr:    errors.New("plot with coordinate x and y is already occupied"),
		},
		{
			name: "Estate Not Found",
			prepareMocks: func(mockRepo *repository.MockRepositoryInterface, mock sqlmock.Sqlmock) {
				mockRepo.EXPECT().GetPlotByXAndY(gomock.Any(), mockEstateID, 1, 2).Return(nil, errors.New("not found"))
				mockRepo.EXPECT().GetEstate(gomock.Any(), mockEstateID).Return(repository.EstateEntity{}, gorm.ErrRecordNotFound)
			},
			request: generated.TreeRequest{
				X:      1,
				Y:      2,
				Height: 10,
			},
			expectedResp:   generated.TreeResponse{},
			expectedStatus: http.StatusNotFound,
			expectedErr:    gorm.ErrRecordNotFound,
		},
		{
			name: "Coordinates Out of Range",
			prepareMocks: func(mockRepo *repository.MockRepositoryInterface, mock sqlmock.Sqlmock) {
				mockEstate := repository.EstateEntity{
					ID:               mockEstateID,
					Length:           5,
					Width:            10,
					TotalDistance:    20,
					TreeMinHeight:    5,
					TreeMaxHeight:    15,
					TreeCount:        1,
					TreeMedianHeight: 10,
				}
				mockRepo.EXPECT().GetPlotByXAndY(gomock.Any(), mockEstateID, 6, 11).Return(nil, errors.New("not found"))
				mockRepo.EXPECT().GetEstate(gomock.Any(), mockEstateID).Return(mockEstate, nil)
			},
			request: generated.TreeRequest{
				X:      6,
				Y:      11,
				Height: 10,
			},
			expectedResp:   generated.TreeResponse{},
			expectedStatus: http.StatusBadRequest,
			expectedErr:    errors.New("x or y is out of range"),
		},
		{
			name: "Occupied Plot Forward Error",
			prepareMocks: func(mockRepo *repository.MockRepositoryInterface, mock sqlmock.Sqlmock) {
				mockEstate := repository.EstateEntity{
					ID:               mockEstateID,
					Length:           5,
					Width:            10,
					TotalDistance:    20,
					TreeMinHeight:    5,
					TreeMaxHeight:    15,
					TreeCount:        1,
					TreeMedianHeight: 10,
				}
				mockRepo.EXPECT().GetPlotByXAndY(gomock.Any(), mockEstateID, 1, 2).Return(nil, errors.New("not found"))
				mockRepo.EXPECT().GetEstate(gomock.Any(), mockEstateID).Return(mockEstate, nil)
				mockRepo.EXPECT().PostPlot(gomock.Any(), gomock.Any()).Return(&mockUUID, nil)
				mockRepo.EXPECT().GetOccupiedPlotBehind(gomock.Any(), mockEstateID, 10).Return(nil, gorm.ErrRecordNotFound)
				mockRepo.EXPECT().GetOccupiedPlotForward(gomock.Any(), mockEstateID, 10).Return(nil, errors.New("some error"))
			},
			request: generated.TreeRequest{
				X:      1,
				Y:      2,
				Height: 10,
			},
			expectedResp:   generated.TreeResponse{},
			expectedStatus: http.StatusInternalServerError,
			expectedErr:    errors.New("some error"),
		},
		{
			name: "Rollback on Repository SaveEstate",
			prepareMocks: func(mockRepo *repository.MockRepositoryInterface, mock sqlmock.Sqlmock) {
				mockEstate := repository.EstateEntity{
					ID:               mockEstateID,
					Length:           5,
					Width:            10,
					TotalDistance:    20,
					TreeMinHeight:    5,
					TreeMaxHeight:    15,
					TreeCount:        1,
					TreeMedianHeight: 10,
				}
				mockRepo.EXPECT().GetPlotByXAndY(gomock.Any(), mockEstateID, 1, 2).Return(nil, errors.New("not found"))
				mockRepo.EXPECT().GetEstate(gomock.Any(), mockEstateID).Return(mockEstate, nil)
				mockRepo.EXPECT().PostPlot(gomock.Any(), gomock.Any()).Return(&mockUUID, nil)
				mockRepo.EXPECT().GetOccupiedPlotBehind(gomock.Any(), mockEstateID, 10).Return(nil, gorm.ErrRecordNotFound)
				mockRepo.EXPECT().GetOccupiedPlotForward(gomock.Any(), mockEstateID, 10).Return(nil, gorm.ErrRecordNotFound)
				mockRepo.EXPECT().GetMedianTreeHeight(gomock.Any(), mockEstateID).Return(10, nil)
				mockRepo.EXPECT().GetPlotByOrderNumber(gomock.Any(), gomock.Any(), gomock.Any()).Return(&repository.PlotEntity{ID: mockUUID}, nil).AnyTimes()

				// Set up the expectation for SaveEstate to return an error
				mockRepo.EXPECT().SaveEstate(gomock.Any(), gomock.Any()).Return(nil, errors.New("save estate failed"))
			},
			request: generated.TreeRequest{
				X:      1,
				Y:      2,
				Height: 10,
			},
			expectedResp:   generated.TreeResponse{},
			expectedStatus: http.StatusInternalServerError,
			expectedErr:    errors.New("save estate failed"),
		},
		{
			name: "Panic Handling",
			prepareMocks: func(mockRepo *repository.MockRepositoryInterface, mock sqlmock.Sqlmock) {
				mockRepo.EXPECT().GetPlotByXAndY(gomock.Any(), mockEstateID, 1, 2).DoAndReturn(func(ctx interface{}, estateID uuid.UUID, x, y int) (*repository.PlotEntity, error) {
					panic("test panic")
				})
			},
			isPanicTest: true,
			request: generated.TreeRequest{
				X:      1,
				Y:      2,
				Height: 10,
			},
			expectedResp:   generated.TreeResponse{},
			expectedStatus: http.StatusInternalServerError,
			expectedErr:    errors.New("test panic"),
		},
		{
			name: "Repository PostPlot Error",
			prepareMocks: func(mockRepo *repository.MockRepositoryInterface, mock sqlmock.Sqlmock) {
				mockRepo.EXPECT().GetPlotByXAndY(gomock.Any(), mockEstateID, 1, 2).Return(nil, errors.New("not found"))
				mockRepo.EXPECT().GetEstate(gomock.Any(), mockEstateID).Return(repository.EstateEntity{}, nil)
			},
			request: generated.TreeRequest{
				X:      1,
				Y:      2,
				Height: 10,
			},
			expectedResp:   generated.TreeResponse{},
			expectedStatus: http.StatusBadRequest,
			expectedErr:    errors.New("x or y is out of range"),
		},
		{
			name: "Successful AddTreeToEstate",
			prepareMocks: func(mockRepo *repository.MockRepositoryInterface, mock sqlmock.Sqlmock) {
				mockPlot := repository.PlotEntity{
					ID:          uuid.New(),
					EstateId:    mockEstateID,
					X:           1,
					Y:           2,
					TreeHeight:  10,
					OrderNumber: 1,
					Distance:    20,
					CreatedAt:   mockTime,
				}
				mockEstate := repository.EstateEntity{
					ID:               mockEstateID,
					Length:           5,
					Width:            10,
					TotalDistance:    20,
					TreeMinHeight:    5,
					TreeMaxHeight:    15,
					TreeCount:        1,
					TreeMedianHeight: 10,
				}
				mockRepo.EXPECT().GetPlotByXAndY(gomock.Any(), mockEstateID, 1, 2).Return(nil, errors.New("not found"))
				mockRepo.EXPECT().GetEstate(gomock.Any(), mockEstateID).Return(mockEstate, nil)
				mockRepo.EXPECT().PostPlot(gomock.Any(), gomock.Any()).Return(&mockUUID, errors.New("some error"))
				mockRepo.EXPECT().GetOccupiedPlotBehind(gomock.Any(), mockEstateID, 10).Return(nil, gorm.ErrRecordNotFound)
				mockRepo.EXPECT().GetPlotByOrderNumber(gomock.Any(), gomock.Any(), gomock.Any()).Return(&repository.PlotEntity{ID: mockPlot.ID}, nil).AnyTimes()

				rows := sqlmock.NewRows([]string{"id"}).AddRow(mockPlot.ID)
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(
					`INSERT INTO "plots" ("estate_id","x","y","distance","order_number","tree_height","created_at") VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING "id"`)).
					WithArgs(mockPlot.EstateId, mockPlot.X, mockPlot.Y, mockPlot.Distance, mockPlot.OrderNumber, mockPlot.TreeHeight, mockPlot.CreatedAt).
					WillReturnRows(rows)
				mock.ExpectCommit()
			},
			request: generated.TreeRequest{
				X:      1,
				Y:      2,
				Height: 10,
			},
			expectedResp: generated.TreeResponse{
				Id: nil,
			},
			expectedStatus: http.StatusInternalServerError,
			expectedErr:    errors.New("some error"),
		},
		{
			name: "Panic Handling with Rollback",
			prepareMocks: func(mockRepo *repository.MockRepositoryInterface, mock sqlmock.Sqlmock) {
				mockRepo.EXPECT().GetPlotByXAndY(gomock.Any(), mockEstateID, 1, 2).DoAndReturn(func(ctx interface{}, estateID uuid.UUID, x, y int) (*repository.PlotEntity, error) {
					panic("test panic")
				})
				mock.ExpectBegin()
				mock.ExpectRollback()
			},
			isPanicTest: true,
			request: generated.TreeRequest{
				X:      1,
				Y:      2,
				Height: 10,
			},
			expectedResp:   generated.TreeResponse{},
			expectedStatus: http.StatusInternalServerError,
			expectedErr:    errors.New("test panic"),
		},
		{
			name: "Plot Already Exists",
			prepareMocks: func(mockRepo *repository.MockRepositoryInterface, mock sqlmock.Sqlmock) {
				mockRepo.EXPECT().GetPlotByXAndY(gomock.Any(), mockEstateID, 1, 2).Return(&mockUUID, nil)
			},
			request: generated.TreeRequest{
				X:      1,
				Y:      2,
				Height: 10,
			},
			expectedResp:   generated.TreeResponse{},
			expectedStatus: http.StatusBadRequest,
			expectedErr:    errors.New("plot with coordinate x and y is already occupied"),
		},
		{
			name: "Estate Not Found",
			prepareMocks: func(mockRepo *repository.MockRepositoryInterface, mock sqlmock.Sqlmock) {
				mockRepo.EXPECT().GetPlotByXAndY(gomock.Any(), mockEstateID, 1, 2).Return(nil, errors.New("not found"))
				mockRepo.EXPECT().GetEstate(gomock.Any(), mockEstateID).Return(repository.EstateEntity{}, errors.New("not found"))
			},
			request: generated.TreeRequest{
				X:      1,
				Y:      2,
				Height: 10,
			},
			expectedResp:   generated.TreeResponse{},
			expectedStatus: http.StatusNotFound,
			expectedErr:    errors.New("not found"),
		},
		{
			name: "Coordinates Out of Range",
			prepareMocks: func(mockRepo *repository.MockRepositoryInterface, mock sqlmock.Sqlmock) {
				mockEstate := repository.EstateEntity{
					ID:               mockEstateID,
					Length:           5,
					Width:            10,
					TotalDistance:    20,
					TreeMinHeight:    5,
					TreeMaxHeight:    15,
					TreeCount:        1,
					TreeMedianHeight: 10,
				}
				mockRepo.EXPECT().GetPlotByXAndY(gomock.Any(), mockEstateID, 6, 11).Return(nil, errors.New("not found"))
				mockRepo.EXPECT().GetEstate(gomock.Any(), mockEstateID).Return(mockEstate, nil)
			},
			request: generated.TreeRequest{
				X:      6,
				Y:      11,
				Height: 10,
			},
			expectedResp:   generated.TreeResponse{},
			expectedStatus: http.StatusBadRequest,
			expectedErr:    errors.New("x or y is out of range"),
		},
		{
			name: "Y is Odd",
			prepareMocks: func(mockRepo *repository.MockRepositoryInterface, mock sqlmock.Sqlmock) {
				mockEstate := repository.EstateEntity{
					ID:               mockEstateID,
					Length:           5,
					Width:            10,
					TotalDistance:    20,
					TreeMinHeight:    5,
					TreeMaxHeight:    15,
					TreeCount:        1,
					TreeMedianHeight: 10,
				}
				mockRepo.EXPECT().GetPlotByXAndY(gomock.Any(), mockEstateID, 3, 5).Return(nil, errors.New("not found"))
				mockRepo.EXPECT().GetEstate(gomock.Any(), mockEstateID).Return(mockEstate, nil)
				mockRepo.EXPECT().PostPlot(gomock.Any(), gomock.Any()).Return(&mockUUID, nil)
				mockRepo.EXPECT().GetOccupiedPlotBehind(gomock.Any(), mockEstateID, 23).Return(nil, gorm.ErrRecordNotFound)
				mockRepo.EXPECT().GetOccupiedPlotForward(gomock.Any(), mockEstateID, 23).Return(nil, gorm.ErrRecordNotFound)
				mockRepo.EXPECT().GetPlotByOrderNumber(gomock.Any(), gomock.Any(), gomock.Any()).Return(&repository.PlotEntity{ID: mockUUID}, nil).AnyTimes()
				mockRepo.EXPECT().GetMedianTreeHeight(gomock.Any(), mockEstateID).Return(10, nil)
				mockRepo.EXPECT().SaveEstate(gomock.Any(), gomock.Any()).Return(nil, nil)
			},
			request: generated.TreeRequest{
				X:      3,
				Y:      5,
				Height: 10,
			},
			expectedResp: generated.TreeResponse{
				Id: &mockUUID,
			},
			expectedStatus: http.StatusOK,
			expectedErr:    nil,
		},
		{
			name: "Occupied Plot Behind Error",
			prepareMocks: func(mockRepo *repository.MockRepositoryInterface, mock sqlmock.Sqlmock) {
				mockEstate := repository.EstateEntity{
					ID:               mockEstateID,
					Length:           5,
					Width:            10,
					TotalDistance:    20,
					TreeMinHeight:    5,
					TreeMaxHeight:    15,
					TreeCount:        1,
					TreeMedianHeight: 10,
				}
				mockRepo.EXPECT().GetPlotByXAndY(gomock.Any(), mockEstateID, 3, 4).Return(nil, errors.New("not found"))
				mockRepo.EXPECT().GetEstate(gomock.Any(), mockEstateID).Return(mockEstate, nil)
				mockRepo.EXPECT().GetOccupiedPlotBehind(gomock.Any(), mockEstateID, 18).Return(nil, errors.New("some other error"))
			},
			request: generated.TreeRequest{
				X:      3,
				Y:      4,
				Height: 10,
			},
			expectedResp:   generated.TreeResponse{},
			expectedStatus: http.StatusInternalServerError,
			expectedErr:    errors.New("some other error"),
		},
		{
			name: "Occupied Plot Behind Exists, Distance Between Plots > 1",
			prepareMocks: func(mockRepo *repository.MockRepositoryInterface, mock sqlmock.Sqlmock) {
				mockOccupiedPlot := repository.PlotEntity{
					ID:          uuid.New(),
					EstateId:    mockEstateID,
					X:           3,
					Y:           4,
					TreeHeight:  12,
					OrderNumber: 14,
					Distance:    120,
				}
				mockEstate := repository.EstateEntity{
					ID:               mockEstateID,
					Length:           5,
					Width:            10,
					TotalDistance:    20,
					TreeMinHeight:    5,
					TreeMaxHeight:    15,
					TreeCount:        1,
					TreeMedianHeight: 10,
				}
				mockRepo.EXPECT().GetPlotByXAndY(gomock.Any(), mockEstateID, 3, 5).Return(nil, errors.New("not found"))
				mockRepo.EXPECT().GetEstate(gomock.Any(), mockEstateID).Return(mockEstate, nil)
				mockRepo.EXPECT().PostPlot(gomock.Any(), gomock.Any()).Return(&mockUUID, nil)
				mockRepo.EXPECT().GetOccupiedPlotBehind(gomock.Any(), mockEstateID, 23).Return(&mockOccupiedPlot, nil)
				mockRepo.EXPECT().GetOccupiedPlotForward(gomock.Any(), mockEstateID, 23).Return(nil, gorm.ErrRecordNotFound)
				mockRepo.EXPECT().GetPlotByOrderNumber(gomock.Any(), gomock.Any(), gomock.Any()).Return(&repository.PlotEntity{ID: mockUUID}, nil).AnyTimes()
				mockRepo.EXPECT().GetMedianTreeHeight(gomock.Any(), mockEstateID).Return(10, nil)
				mockRepo.EXPECT().SaveEstate(gomock.Any(), gomock.Any()).Return(nil, nil)
			},
			request: generated.TreeRequest{
				X:      3,
				Y:      5,
				Height: 10,
			},
			expectedResp: generated.TreeResponse{
				Id: &mockUUID,
			},
			expectedStatus: http.StatusOK,
			expectedErr:    nil,
		},
		{
			name: "Occupied Plot Behind Exists, Distance Between Plots is 1",
			prepareMocks: func(mockRepo *repository.MockRepositoryInterface, mock sqlmock.Sqlmock) {
				mockOccupiedPlot := repository.PlotEntity{
					ID:          uuid.New(),
					EstateId:    mockEstateID,
					X:           3,
					Y:           4,
					TreeHeight:  12,
					OrderNumber: 22,
					Distance:    140,
				}
				mockEstate := repository.EstateEntity{
					ID:               mockEstateID,
					Length:           5,
					Width:            10,
					TotalDistance:    20,
					TreeMinHeight:    5,
					TreeMaxHeight:    15,
					TreeCount:        1,
					TreeMedianHeight: 10,
				}
				mockRepo.EXPECT().GetPlotByXAndY(gomock.Any(), mockEstateID, 3, 5).Return(nil, errors.New("not found"))
				mockRepo.EXPECT().GetEstate(gomock.Any(), mockEstateID).Return(mockEstate, nil)
				mockRepo.EXPECT().PostPlot(gomock.Any(), gomock.Any()).Return(&mockUUID, nil)
				mockRepo.EXPECT().GetOccupiedPlotBehind(gomock.Any(), mockEstateID, 23).Return(&mockOccupiedPlot, nil)
				mockRepo.EXPECT().GetOccupiedPlotForward(gomock.Any(), mockEstateID, 23).Return(nil, gorm.ErrRecordNotFound)
				mockRepo.EXPECT().GetPlotByOrderNumber(gomock.Any(), gomock.Any(), gomock.Any()).Return(&repository.PlotEntity{ID: mockUUID}, nil).AnyTimes()
				mockRepo.EXPECT().GetMedianTreeHeight(gomock.Any(), mockEstateID).Return(10, nil)
				mockRepo.EXPECT().SaveEstate(gomock.Any(), gomock.Any()).Return(nil, nil)
			},
			request: generated.TreeRequest{
				X:      3,
				Y:      5,
				Height: 10,
			},
			expectedResp: generated.TreeResponse{
				Id: &mockUUID,
			},
			expectedStatus: http.StatusOK,
			expectedErr:    nil,
		},
		{
			name: "Occupied Plot Forward Exists, Error Returned",
			prepareMocks: func(mockRepo *repository.MockRepositoryInterface, mock sqlmock.Sqlmock) {
				mockOccupiedPlotBehind := repository.PlotEntity{
					ID:          uuid.New(),
					EstateId:    mockEstateID,
					X:           3,
					Y:           4,
					TreeHeight:  12,
					OrderNumber: 22,
					Distance:    140,
				}
				mockOccupiedPlotForward := repository.PlotEntity{
					ID:          uuid.New(),
					EstateId:    mockEstateID,
					X:           3,
					Y:           6,
					TreeHeight:  15,
					OrderNumber: 24,
					Distance:    160,
				}
				mockEstate := repository.EstateEntity{
					ID:               mockEstateID,
					Length:           5,
					Width:            10,
					TotalDistance:    20,
					TreeMinHeight:    5,
					TreeMaxHeight:    15,
					TreeCount:        1,
					TreeMedianHeight: 10,
				}
				mockRepo.EXPECT().GetPlotByXAndY(gomock.Any(), mockEstateID, 3, 5).Return(nil, errors.New("not found"))
				mockRepo.EXPECT().GetEstate(gomock.Any(), mockEstateID).Return(mockEstate, nil)
				mockRepo.EXPECT().PostPlot(gomock.Any(), gomock.Any()).Return(&mockUUID, nil)
				mockRepo.EXPECT().GetOccupiedPlotBehind(gomock.Any(), mockEstateID, 23).Return(&mockOccupiedPlotBehind, nil)
				mockRepo.EXPECT().GetOccupiedPlotForward(gomock.Any(), mockEstateID, 23).Return(&mockOccupiedPlotForward, errors.New("database error"))
				mockRepo.EXPECT().GetPlotByOrderNumber(gomock.Any(), gomock.Any(), gomock.Any()).Return(&repository.PlotEntity{ID: mockUUID}, nil).AnyTimes()
			},
			request: generated.TreeRequest{
				X:      3,
				Y:      5,
				Height: 10,
			},
			expectedResp: generated.TreeResponse{
				Id: nil,
			},
			expectedStatus: http.StatusInternalServerError,
			expectedErr:    errors.New("database error"),
		},
		{
			name: "Occupied Plot Forward not error, but distance is 1",
			prepareMocks: func(mockRepo *repository.MockRepositoryInterface, mock sqlmock.Sqlmock) {
				mockOccupiedPlotBehind := repository.PlotEntity{
					ID:          uuid.New(),
					EstateId:    mockEstateID,
					X:           3,
					Y:           4,
					TreeHeight:  12,
					OrderNumber: 22,
					Distance:    140,
				}
				mockOccupiedPlotForward := repository.PlotEntity{
					ID:          uuid.New(),
					EstateId:    mockEstateID,
					X:           3,
					Y:           6,
					TreeHeight:  15,
					OrderNumber: 24,
					Distance:    167,
				}
				mockEstate := repository.EstateEntity{
					ID:               mockEstateID,
					Length:           5,
					Width:            10,
					TotalDistance:    20,
					TreeMinHeight:    5,
					TreeMaxHeight:    15,
					TreeCount:        1,
					TreeMedianHeight: 10,
				}
				mockRepo.EXPECT().GetPlotByXAndY(gomock.Any(), mockEstateID, 3, 5).Return(nil, errors.New("not found"))
				mockRepo.EXPECT().GetEstate(gomock.Any(), mockEstateID).Return(mockEstate, nil)
				mockRepo.EXPECT().PostPlot(gomock.Any(), gomock.Any()).Return(&mockUUID, nil)
				mockRepo.EXPECT().GetOccupiedPlotBehind(gomock.Any(), mockEstateID, 23).Return(&mockOccupiedPlotBehind, nil)
				mockRepo.EXPECT().GetOccupiedPlotForward(gomock.Any(), mockEstateID, 23).Return(&mockOccupiedPlotForward, nil)
				mockRepo.EXPECT().GetPlotByOrderNumber(gomock.Any(), gomock.Any(), gomock.Any()).Return(&repository.PlotEntity{ID: mockUUID}, nil).AnyTimes()
				mockRepo.EXPECT().SavePlot(gomock.Any(), mockOccupiedPlotForward).Return(nil, errors.New("database error"))

			},
			request: generated.TreeRequest{
				X:      3,
				Y:      5,
				Height: 10,
			},
			expectedResp: generated.TreeResponse{
				Id: nil,
			},
			expectedStatus: http.StatusInternalServerError,
			expectedErr:    errors.New("database error"),
		},
		{
			name: "Occupied Plot Forward not error, but distance is > 1",
			prepareMocks: func(mockRepo *repository.MockRepositoryInterface, mock sqlmock.Sqlmock) {
				mockOccupiedPlotBehind := repository.PlotEntity{
					ID:          uuid.New(),
					EstateId:    mockEstateID,
					X:           3,
					Y:           4,
					TreeHeight:  12,
					OrderNumber: 22,
					Distance:    140,
				}
				mockOccupiedPlotForward := repository.PlotEntity{
					ID:          uuid.New(),
					EstateId:    mockEstateID,
					X:           3,
					Y:           6,
					TreeHeight:  15,
					OrderNumber: 29,
					Distance:    239,
				}
				mockEstate := repository.EstateEntity{
					ID:               mockEstateID,
					Length:           5,
					Width:            10,
					TotalDistance:    20,
					TreeMinHeight:    5,
					TreeMaxHeight:    15,
					TreeCount:        1,
					TreeMedianHeight: 10,
				}
				mockRepo.EXPECT().GetPlotByXAndY(gomock.Any(), mockEstateID, 3, 5).Return(nil, errors.New("not found"))
				mockRepo.EXPECT().GetEstate(gomock.Any(), mockEstateID).Return(mockEstate, nil)
				mockRepo.EXPECT().PostPlot(gomock.Any(), gomock.Any()).Return(&mockUUID, nil)
				mockRepo.EXPECT().GetOccupiedPlotBehind(gomock.Any(), mockEstateID, 23).Return(&mockOccupiedPlotBehind, nil)
				mockRepo.EXPECT().GetOccupiedPlotForward(gomock.Any(), mockEstateID, 23).Return(&mockOccupiedPlotForward, nil)
				mockRepo.EXPECT().GetPlotByOrderNumber(gomock.Any(), gomock.Any(), gomock.Any()).Return(&repository.PlotEntity{ID: mockUUID}, nil).AnyTimes()
				mockRepo.EXPECT().SavePlot(gomock.Any(), mockOccupiedPlotForward).Return(nil, errors.New("database error"))

			},
			request: generated.TreeRequest{
				X:      3,
				Y:      5,
				Height: 10,
			},
			expectedResp: generated.TreeResponse{
				Id: nil,
			},
			expectedStatus: http.StatusInternalServerError,
			expectedErr:    errors.New("database error"),
		},
		{
			name: "AdjustPlotForwardDistance Error",
			prepareMocks: func(mockRepo *repository.MockRepositoryInterface, mock sqlmock.Sqlmock) {
				mockOccupiedPlotBehind := repository.PlotEntity{
					ID:          uuid.New(),
					EstateId:    mockEstateID,
					X:           3,
					Y:           4,
					TreeHeight:  12,
					OrderNumber: 22,
					Distance:    140,
				}
				mockOccupiedPlotForward := repository.PlotEntity{
					ID:          uuid.New(),
					EstateId:    mockEstateID,
					X:           3,
					Y:           6,
					TreeHeight:  15,
					OrderNumber: 23,
					Distance:    179,
				}
				mockEstate := repository.EstateEntity{
					ID:               mockEstateID,
					Length:           5,
					Width:            10,
					TotalDistance:    20,
					TreeMinHeight:    5,
					TreeMaxHeight:    15,
					TreeCount:        1,
					TreeMedianHeight: 10,
				}
				mockRepo.EXPECT().GetPlotByXAndY(gomock.Any(), mockEstateID, 3, 5).Return(nil, errors.New("not found"))
				mockRepo.EXPECT().GetEstate(gomock.Any(), mockEstateID).Return(mockEstate, nil)
				mockRepo.EXPECT().PostPlot(gomock.Any(), gomock.Any()).Return(&mockUUID, nil)
				mockRepo.EXPECT().GetOccupiedPlotBehind(gomock.Any(), mockEstateID, 23).Return(&mockOccupiedPlotBehind, nil)
				mockRepo.EXPECT().GetOccupiedPlotForward(gomock.Any(), mockEstateID, 23).Return(&mockOccupiedPlotForward, nil)
				mockRepo.EXPECT().GetPlotByOrderNumber(gomock.Any(), gomock.Any(), gomock.Any()).Return(&repository.PlotEntity{ID: mockUUID}, nil).AnyTimes()
				mockRepo.EXPECT().SavePlot(gomock.Any(), mockOccupiedPlotForward).Return(nil, nil)
				mockRepo.EXPECT().AdjustPlotForwardDistance(gomock.Any(), mockEstateID, 23, gomock.Any()).Return(errors.New("adjust distance error"))
			},
			request: generated.TreeRequest{
				X:      3,
				Y:      5,
				Height: 10,
			},
			expectedResp:   generated.TreeResponse{}, // Empty response due to error
			expectedStatus: http.StatusInternalServerError,
			expectedErr:    errors.New("adjust distance error"),
		},
		{
			name: "Error Fetching Median Tree Height",
			prepareMocks: func(mockRepo *repository.MockRepositoryInterface, mock sqlmock.Sqlmock) {
				mockRepo.EXPECT().GetPlotByXAndY(gomock.Any(), mockEstateID, 1, 2).Return(nil, errors.New("not found"))
				mockRepo.EXPECT().GetEstate(gomock.Any(), mockEstateID).Return(repository.EstateEntity{
					ID:               mockEstateID,
					Length:           5,
					Width:            10,
					TotalDistance:    20,
					TreeMinHeight:    0,
					TreeMaxHeight:    15,
					TreeCount:        1,
					TreeMedianHeight: 10,
				}, nil)
				mockRepo.EXPECT().GetMedianTreeHeight(gomock.Any(), mockEstateID).Return(0, errors.New("database error"))
				mockRepo.EXPECT().PostPlot(gomock.Any(), gomock.Any()).Return(&mockUUID, nil)
				mockRepo.EXPECT().GetOccupiedPlotBehind(gomock.Any(), mockEstateID, 10).Return(nil, gorm.ErrRecordNotFound)
				mockRepo.EXPECT().GetOccupiedPlotForward(gomock.Any(), mockEstateID, 10).Return(nil, gorm.ErrRecordNotFound)
			},
			request: generated.TreeRequest{
				X:      1,
				Y:      2,
				Height: 10,
			},
			expectedResp:   generated.TreeResponse{},
			expectedStatus: http.StatusInternalServerError,
			expectedErr:    errors.New("database error"),
		},
		{
			name: "Error Fetching Plot by Order Number previous tree",
			prepareMocks: func(mockRepo *repository.MockRepositoryInterface, mock sqlmock.Sqlmock) {
				mockRepo.EXPECT().GetPlotByXAndY(gomock.Any(), mockEstateID, 1, 2).Return(nil, errors.New("not found"))
				mockRepo.EXPECT().GetEstate(gomock.Any(), mockEstateID).Return(repository.EstateEntity{
					ID:               mockEstateID,
					Length:           5,
					Width:            10,
					TotalDistance:    20,
					TreeMinHeight:    0,
					TreeMaxHeight:    15,
					TreeCount:        1,
					TreeMedianHeight: 10,
				}, nil)
				mockRepo.EXPECT().GetMedianTreeHeight(gomock.Any(), mockEstateID).Return(10, nil)
				mockRepo.EXPECT().PostPlot(gomock.Any(), gomock.Any()).Return(&mockUUID, nil)
				mockRepo.EXPECT().GetPlotByOrderNumber(gomock.Any(), mockEstateID, 9).Return(nil, errors.New("plot by order number error"))
				mockRepo.EXPECT().GetOccupiedPlotBehind(gomock.Any(), mockEstateID, 10).Return(nil, gorm.ErrRecordNotFound)
				mockRepo.EXPECT().GetOccupiedPlotForward(gomock.Any(), mockEstateID, 10).Return(nil, gorm.ErrRecordNotFound)
			},
			request: generated.TreeRequest{
				X:      1,
				Y:      2,
				Height: 10,
			},
			expectedResp:   generated.TreeResponse{},
			expectedStatus: http.StatusInternalServerError,
			expectedErr:    errors.New("plot by order number error"),
		},
		{
			name: "Error Fetching Plot by Order Number forward tree",
			prepareMocks: func(mockRepo *repository.MockRepositoryInterface, mock sqlmock.Sqlmock) {
				prevPlot := repository.PlotEntity{
					ID:          uuid.New(),
					EstateId:    mockEstateID,
					X:           3,
					Y:           4,
					TreeHeight:  12,
					OrderNumber: 22,
					Distance:    140,
				}
				mockRepo.EXPECT().GetPlotByXAndY(gomock.Any(), mockEstateID, 1, 2).Return(nil, errors.New("not found"))
				mockRepo.EXPECT().GetEstate(gomock.Any(), mockEstateID).Return(repository.EstateEntity{
					ID:               mockEstateID,
					Length:           5,
					Width:            10,
					TotalDistance:    20,
					TreeMinHeight:    0,
					TreeMaxHeight:    15,
					TreeCount:        1,
					TreeMedianHeight: 10,
				}, nil)
				mockRepo.EXPECT().GetMedianTreeHeight(gomock.Any(), mockEstateID).Return(10, nil)
				mockRepo.EXPECT().PostPlot(gomock.Any(), gomock.Any()).Return(&mockUUID, nil)
				mockRepo.EXPECT().GetPlotByOrderNumber(gomock.Any(), mockEstateID, 9).Return(&prevPlot, nil)
				mockRepo.EXPECT().GetPlotByOrderNumber(gomock.Any(), mockEstateID, 11).Return(nil, errors.New("plot by order number error"))
				mockRepo.EXPECT().GetOccupiedPlotBehind(gomock.Any(), mockEstateID, 10).Return(nil, gorm.ErrRecordNotFound)
				mockRepo.EXPECT().GetOccupiedPlotForward(gomock.Any(), mockEstateID, 10).Return(nil, gorm.ErrRecordNotFound)
			},
			request: generated.TreeRequest{
				X:      1,
				Y:      2,
				Height: 10,
			},
			expectedResp:   generated.TreeResponse{},
			expectedStatus: http.StatusInternalServerError,
			expectedErr:    errors.New("plot by order number error"),
		},
		{
			name: "Error Fetching Plot by Order Number forward tree",
			prepareMocks: func(mockRepo *repository.MockRepositoryInterface, mock sqlmock.Sqlmock) {
				mockRepo.EXPECT().GetPlotByXAndY(gomock.Any(), mockEstateID, 1, 2).Return(nil, errors.New("not found"))
				mockRepo.EXPECT().GetEstate(gomock.Any(), mockEstateID).Return(repository.EstateEntity{
					ID:               mockEstateID,
					Length:           5,
					Width:            10,
					TotalDistance:    20,
					TreeMinHeight:    0,
					TreeMaxHeight:    15,
					TreeCount:        1,
					TreeMedianHeight: 10,
				}, nil)
				mockRepo.EXPECT().GetMedianTreeHeight(gomock.Any(), mockEstateID).Return(10, nil)
				mockRepo.EXPECT().PostPlot(gomock.Any(), gomock.Any()).Return(&mockUUID, nil)
				mockRepo.EXPECT().GetPlotByOrderNumber(gomock.Any(), mockEstateID, 9).Return(nil, nil)
				mockRepo.EXPECT().GetPlotByOrderNumber(gomock.Any(), mockEstateID, 11).Return(nil, nil)
				mockRepo.EXPECT().GetOccupiedPlotBehind(gomock.Any(), mockEstateID, 10).Return(nil, gorm.ErrRecordNotFound)
				mockRepo.EXPECT().GetOccupiedPlotForward(gomock.Any(), mockEstateID, 10).Return(nil, gorm.ErrRecordNotFound)
				mockRepo.EXPECT().SaveEstate(gomock.Any(), gomock.Any()).Return(nil, errors.New("save estate failed"))
			},
			request: generated.TreeRequest{
				X:      1,
				Y:      2,
				Height: 10,
			},
			expectedResp:   generated.TreeResponse{},
			expectedStatus: http.StatusInternalServerError,
			expectedErr:    errors.New("save estate failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)

			gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
			require.NoError(t, err)

			mockRepo := repository.NewMockRepositoryInterface(ctrl)
			tt.prepareMocks(mockRepo, mock)

			service := &Service{
				Repository: mockRepo,
				Db:         gdb,
			}

			e := echo.New()

			jsonReq, err := json.Marshal(tt.request)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(jsonReq))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			defer func() {
				if r := recover(); r != nil {
					require.Equal(t, tt.expectedErr.Error(), r)
				}
			}()

			resp, status, err := service.AddTreeToEstate(c, tt.request, mockEstateID)

			if tt.isPanicTest {
				defer func() {
					if r := recover(); r != nil {
						require.Equal(t, tt.expectedErr.Error(), r)
						err := mock.ExpectationsWereMet()
						if err != nil {
							return
						} // Ensure all expectations are met, including rollback
					}
				}()
			} else if tt.expectedErr != nil {
				require.Error(t, err)
				require.Equal(t, tt.expectedErr.Error(), err.Error())
				require.Equal(t, tt.expectedResp, resp)
				require.Equal(t, tt.expectedStatus, status)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedResp, resp)
				require.Equal(t, tt.expectedStatus, status)
			}
		})
	}
}
