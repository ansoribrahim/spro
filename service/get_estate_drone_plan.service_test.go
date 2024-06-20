package service

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"spgo/generated"
	"spgo/repository"
)

func TestService_GetEstateDronePlan(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEstateID := uuid.New()
	mockMaxDistance := 200
	mockContext := context.TODO()
	mockX := 9 // Update as per your expected behavior
	mockY := 5 // Update as per your expected behavior
	respDistance := 200

	tests := []struct {
		name         string
		prepareMocks func(mockRepo *repository.MockRepositoryInterface, mock sqlmock.Sqlmock)
		estateID     uuid.UUID
		maxDistance  *int
		expectedResp generated.DronePlanResponse
		expectedErr  error
	}{
		{
			name: "Successful Scenario",
			prepareMocks: func(mockRepo *repository.MockRepositoryInterface, mock sqlmock.Sqlmock) {
				mockEstate := repository.EstateEntity{
					ID:            mockEstateID,
					Length:        10,
					TotalDistance: 200,
				}
				mockPlot := repository.PlotEntity{
					ID:       uuid.New(),
					EstateId: mockEstateID,
					X:        5, // Update with correct mock data
					Y:        5, // Update with correct mock data
					Distance: 150,
				}
				mockRepo.EXPECT().GetEstate(gomock.Any(), mockEstateID).Return(mockEstate, nil)
				mockRepo.EXPECT().GetPlotByDistance(gomock.Any(), mockEstateID, mockMaxDistance-1).Return(&mockPlot, nil)
			},
			estateID:    mockEstateID,
			maxDistance: &mockMaxDistance,
			expectedResp: generated.DronePlanResponse{
				Distance: &respDistance,
				Rest: &struct {
					X *int `json:"x,omitempty"`
					Y *int `json:"y,omitempty"`
				}{
					X: &mockX,
					Y: &mockY,
				},
			},
			expectedErr: nil,
		},
		{
			name: "Estate Not Found",
			prepareMocks: func(mockRepo *repository.MockRepositoryInterface, mock sqlmock.Sqlmock) {
				mockRepo.EXPECT().GetEstate(gomock.Any(), mockEstateID).Return(repository.EstateEntity{}, gorm.ErrRecordNotFound)
			},
			estateID:     mockEstateID,
			maxDistance:  &mockMaxDistance,
			expectedResp: generated.DronePlanResponse{},
			expectedErr:  errors.New("estate not found"),
		},
		{
			name: "Other Repository Error",
			prepareMocks: func(mockRepo *repository.MockRepositoryInterface, mock sqlmock.Sqlmock) {
				mockRepo.EXPECT().GetEstate(gomock.Any(), mockEstateID).Return(repository.EstateEntity{}, errors.New("some repository error"))
			},
			estateID:     mockEstateID,
			maxDistance:  &mockMaxDistance,
			expectedResp: generated.DronePlanResponse{},
			expectedErr:  errors.New("some repository error"),
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
				Db:         gdb,
				Repository: mockRepo,
			}

			resp, err := service.GetEstateDronePlan(mockContext, tt.estateID, tt.maxDistance)

			assert.Equal(t, tt.expectedResp, resp)
			if tt.expectedErr == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.expectedErr.Error())
			}
		})
	}
}
