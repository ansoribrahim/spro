package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"spgo/generated"
	"spgo/repository"
	"spgo/service"
)

func TestService_PostEstate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockContext := context.TODO()
	mockRequest := generated.EstateRequest{
		Width:  5,
		Length: 10,
	}
	mockUUID := uuid.New()

	tests := []struct {
		name         string
		prepareMocks func(mockRepo *repository.MockRepositoryInterface)
		request      generated.EstateRequest
		expectedResp generated.EstateResponse
		expectedErr  error
	}{
		{
			name: "Successful Post",
			prepareMocks: func(mockRepo *repository.MockRepositoryInterface) {
				mockRepo.EXPECT().PostEstate(gomock.Any(), gomock.Any()).Return(&mockUUID, nil)
			},
			request: mockRequest,
			expectedResp: generated.EstateResponse{
				Id: &mockUUID,
			},
			expectedErr: nil,
		},
		{
			name: "Repository Error",
			prepareMocks: func(mockRepo *repository.MockRepositoryInterface) {
				mockRepo.EXPECT().PostEstate(gomock.Any(), gomock.Any()).Return(nil, errors.New("repository error"))
			},
			request:      mockRequest,
			expectedResp: generated.EstateResponse{},
			expectedErr:  errors.New("repository error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := repository.NewMockRepositoryInterface(ctrl)
			tt.prepareMocks(mockRepo)

			svs := service.NewService(service.NewServiceOptions{
				Repository: mockRepo,
			})

			resp, err := svs.PostEstate(mockContext, tt.request)

			assert.Equal(t, tt.expectedResp, resp)
			if tt.expectedErr == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.expectedErr.Error())
			}
		})
	}
}
