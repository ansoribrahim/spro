// service/service_test.go
package service

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"spgo/generated"
	"spgo/repository"
)

func TestService_PostEstate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUUID := uuid.New()

	tests := []struct {
		name         string
		prepareMock  func(mockRepo *repository.MockRepositoryInterface)
		request      generated.EstateRequest
		expectedResp generated.EstateResponse
		expectedErr  error
	}{
		{
			name: "Successful PostEstate",
			prepareMock: func(mockRepo *repository.MockRepositoryInterface) {
				mockRepo.EXPECT().
					PostEstate(gomock.Any(), gomock.Any()).
					Return(&mockUUID, nil)
			},
			request: generated.EstateRequest{
				Width:  5,
				Length: 10,
			},
			expectedResp: generated.EstateResponse{
				Id: &mockUUID,
			},
			expectedErr: nil,
		},
		{
			name: "Repository Error",
			prepareMock: func(mockRepo *repository.MockRepositoryInterface) {
				mockRepo.EXPECT().
					PostEstate(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("repository error"))
			},
			request: generated.EstateRequest{
				Width:  5,
				Length: 10,
			},
			expectedResp: generated.EstateResponse{},
			expectedErr:  errors.New("repository error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := repository.NewMockRepositoryInterface(ctrl)
			tt.prepareMock(mockRepo)

			service := NewService(NewServiceOptions{Repository: mockRepo})

			ctx := context.TODO()
			resp, err := service.PostEstate(ctx, tt.request)

			if tt.expectedErr != nil {
				require.Error(t, err)
				require.Equal(t, tt.expectedErr.Error(), err.Error())
				require.Equal(t, generated.EstateResponse{}, resp)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedResp, resp)
			}
		})
	}
}
