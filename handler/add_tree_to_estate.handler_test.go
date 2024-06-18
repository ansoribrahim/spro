package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"spgo/generated"
	"spgo/handler"
	"spgo/service"
)

func TestAddTreeToEstate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUUID := uuid.New()

	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}

	tests := []struct {
		name           string
		id             openapi_types.UUID
		requestBody    string
		mockResponse   generated.TreeResponse
		mockError      error
		expectedError  *string
		prepareMock    func(mockService *service.MockServiceInterface)
		expectedStatus int
	}{
		{
			name:        "Valid Request",
			id:          mockUUID,
			requestBody: `{"x": 1, "y": 2, "height": 15}`,
			mockResponse: generated.TreeResponse{
				Id: &mockUUID,
			},
			mockError:     nil,
			expectedError: nil,
			prepareMock: func(mockService *service.MockServiceInterface) {
				mockService.EXPECT().AddTreeToEstate(gomock.Any(), gomock.Any(), mockUUID).Return(generated.TreeResponse{Id: &mockUUID}, http.StatusCreated, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Invalid JSON",
			id:             mockUUID,
			requestBody:    `{"height": "invalid"}`,
			mockResponse:   generated.TreeResponse{},
			mockError:      nil,
			expectedError:  ptr("Invalid request"),
			prepareMock:    func(mockService *service.MockServiceInterface) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid Parameters",
			id:             mockUUID,
			requestBody:    `{"height": -5}`,
			mockResponse:   generated.TreeResponse{},
			mockError:      nil,
			expectedError:  ptr("Key: 'TreeRequest.Height' Error:Field validation for 'Height' failed"),
			prepareMock:    func(mockService *service.MockServiceInterface) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:          "Service Error",
			id:            mockUUID,
			requestBody:   `{"x": 1, "y": 2, "height": 15}`,
			mockResponse:  generated.TreeResponse{},
			mockError:     nil,
			expectedError: ptr("service error"),
			prepareMock: func(mockService *service.MockServiceInterface) {
				mockService.EXPECT().AddTreeToEstate(gomock.Any(), gomock.Any(), mockUUID).Return(generated.TreeResponse{}, http.StatusBadRequest, errors.New("service error"))
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Height Exceeds",
			id:             openapi_types.UUID(mockUUID),
			requestBody:    `{"height": 100000}`,
			mockResponse:   generated.TreeResponse{},
			mockError:      nil,
			expectedError:  ptr("Key: 'TreeRequest.Height' Error:Field validation for 'Height' failed"),
			prepareMock:    func(mockService *service.MockServiceInterface) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockService := service.NewMockServiceInterface(ctrl)

			tc.prepareMock(mockService)

			server := handler.NewServer(handler.NewServerOptions{Service: mockService})

			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(tc.requestBody)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tc.id.String())

			err := server.AddTreeToEstate(c, tc.id)

			assert.Equal(t, tc.expectedStatus, rec.Code)

			if tc.expectedStatus == http.StatusCreated {
				var resp generated.TreeResponse
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

			assert.Equal(t, tc.mockError, err)
		})
	}
}
