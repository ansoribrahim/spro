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

func TestGetEstateIdDronePlan(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUUID := uuid.New()

	e := echo.New()

	tests := []struct {
		name           string
		id             openapi_types.UUID
		maxDistance    *int
		mockResponse   generated.DronePlanResponse
		mockError      error
		expectedError  *string
		prepareMock    func(mockService *service.MockServiceInterface)
		expectedStatus int
	}{
		{
			name:        "Valid Request",
			id:          mockUUID,
			maxDistance: ptrInt(100),
			mockResponse: generated.DronePlanResponse{
				Distance: ptrInt(100),
				Rest: &struct {
					X *int `json:"x,omitempty"`
					Y *int `json:"y,omitempty"`
				}{
					X: ptrInt(10),
					Y: ptrInt(20),
				},
			},
			mockError:     nil,
			expectedError: nil,
			prepareMock: func(mockService *service.MockServiceInterface) {
				mockService.EXPECT().GetEstateDronePlan(gomock.Any(), mockUUID, ptrInt(100)).Return(generated.DronePlanResponse{
					Distance: ptrInt(100),
					Rest: &struct {
						X *int `json:"x,omitempty"`
						Y *int `json:"y,omitempty"`
					}{
						X: ptrInt(10),
						Y: ptrInt(20),
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:          "Service Error",
			id:            mockUUID,
			maxDistance:   ptrInt(100),
			mockResponse:  generated.DronePlanResponse{},
			mockError:     nil,
			expectedError: ptr("service error"),
			prepareMock: func(mockService *service.MockServiceInterface) {
				mockService.EXPECT().GetEstateDronePlan(gomock.Any(), mockUUID, ptrInt(100)).Return(generated.DronePlanResponse{}, errors.New("service error"))
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
			q := req.URL.Query()
			if tc.maxDistance != nil {
				q.Add("maxDistance", string(rune(*tc.maxDistance)))
			}
			req.URL.RawQuery = q.Encode()
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tc.id.String())

			params := generated.GetEstateIdDronePlanParams{
				MaxDistance: tc.maxDistance,
			}

			err := server.GetEstateIdDronePlan(c, tc.id, params)

			assert.Equal(t, tc.expectedStatus, rec.Code)

			if tc.expectedStatus == http.StatusOK {
				var resp generated.DronePlanResponse
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

func ptrInt(i int) *int {
	return &i
}
