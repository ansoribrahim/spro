package handler_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"spgo/generated"
	"spgo/handler"
	"spgo/service"
)

func TestGetEstateIdStats(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUUID := uuid.New()

	e := echo.New()

	tests := []struct {
		name           string
		id             openapi_types.UUID
		mockResponse   generated.EstateStatsResponse
		mockError      error
		expectedError  *string
		prepareMock    func(mockService *service.MockServiceInterface)
		expectedStatus int
	}{
		{
			name: "Valid Request",
			id:   mockUUID,
			mockResponse: generated.EstateStatsResponse{
				Count:  ptrInt(100),
				Max:    ptrInt(30),
				Median: ptrInt(15),
				Min:    ptrInt(5),
			},
			mockError:     nil,
			expectedError: nil,
			prepareMock: func(mockService *service.MockServiceInterface) {
				mockService.EXPECT().GetEstateStats(gomock.Any(), mockUUID).Return(generated.EstateStatsResponse{
					Count:  ptrInt(100),
					Max:    ptrInt(30),
					Median: ptrInt(15),
					Min:    ptrInt(5),
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:          "Service Error",
			id:            mockUUID,
			mockResponse:  generated.EstateStatsResponse{},
			mockError:     nil,
			expectedError: ptr("service error"),
			prepareMock: func(mockService *service.MockServiceInterface) {
				mockService.EXPECT().GetEstateStats(gomock.Any(), mockUUID).Return(generated.EstateStatsResponse{}, errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockService := service.NewMockServiceInterface(ctrl)

			tc.prepareMock(mockService)

			server := handler.NewServer(handler.NewServerOptions{Service: mockService})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tc.id.String())

			err := server.GetEstateIdStats(c, tc.id)

			assert.Equal(t, tc.expectedStatus, rec.Code)

			if tc.expectedStatus == http.StatusOK {
				var resp generated.EstateStatsResponse
				err1 := json.Unmarshal(rec.Body.Bytes(), &resp)
				assert.NoError(t, err1)
				assert.Equal(t, tc.mockResponse, resp)
			} else {
				var resp map[string]string
				err2 := json.Unmarshal(rec.Body.Bytes(), &resp)
				assert.NoError(t, err2)
				if tc.expectedError != nil {
					assert.Contains(t, resp["error"], *tc.expectedError)
				}
			}

			if tc.mockError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.mockError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
