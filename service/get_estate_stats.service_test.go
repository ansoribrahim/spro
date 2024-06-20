package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"spgo/generated"
	"spgo/repository"
	"spgo/service"
)

func TestService_GetEstateStats(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEstateID := uuid.New()
	mockContext := context.TODO()

	tests := []struct {
		name         string
		prepareMocks func(mockRepo *repository.MockRepositoryInterface)
		estateID     uuid.UUID
		expectedResp generated.EstateStatsResponse
		expectedErr  error
	}{
		{
			name: "Successful Scenario",
			prepareMocks: func(mockRepo *repository.MockRepositoryInterface) {
				mockEstate := repository.EstateEntity{
					ID:               mockEstateID,
					TreeMaxHeight:    100,
					TreeMinHeight:    50,
					TreeMedianHeight: 75,
					TreeCount:        500,
				}
				mockRepo.EXPECT().GetEstate(gomock.Any(), mockEstateID).Return(mockEstate, nil)
			},
			estateID: mockEstateID,
			expectedResp: generated.EstateStatsResponse{
				Max:    &[]int{100}[0],
				Min:    &[]int{50}[0],
				Median: &[]int{75}[0],
				Count:  &[]int{500}[0],
			},
			expectedErr: nil,
		},
		{
			name: "Estate Not Found",
			prepareMocks: func(mockRepo *repository.MockRepositoryInterface) {
				mockRepo.EXPECT().GetEstate(gomock.Any(), mockEstateID).Return(repository.EstateEntity{}, gorm.ErrRecordNotFound)
			},
			estateID:     mockEstateID,
			expectedResp: generated.EstateStatsResponse{},
			expectedErr:  gorm.ErrRecordNotFound,
		},
		{
			name: "Other Repository Error",
			prepareMocks: func(mockRepo *repository.MockRepositoryInterface) {
				mockRepo.EXPECT().GetEstate(gomock.Any(), mockEstateID).Return(repository.EstateEntity{}, errors.New("some repository error"))
			},
			estateID:     mockEstateID,
			expectedResp: generated.EstateStatsResponse{},
			expectedErr:  errors.New("some repository error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := repository.NewMockRepositoryInterface(ctrl)
			tt.prepareMocks(mockRepo)

			service := &service.Service{
				Repository: mockRepo,
			}

			resp, err := service.GetEstateStats(mockContext, tt.estateID)

			assert.Equal(t, tt.expectedResp, resp)
			if tt.expectedErr == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.expectedErr.Error())
			}
		})
	}
}
