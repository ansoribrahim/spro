// handler_test/estate_handler_test.go
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

	"spgo/generated"
	"spgo/handler"
	"spgo/service"
)

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func TestPostEstateHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockUUID := uuid.New()

	defer ctrl.Finish()

	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}

	tests := []struct {
		name           string
		requestBody    string
		mockResponse   generated.EstateResponse
		mockError      error
		expectedError  *string
		prepareMock    func(mockService *service.MockServiceInterface)
		expectedStatus int
	}{
		{
			name:        "Valid Request",
			requestBody: `{"width": 5, "length": 10}`,
			mockResponse: generated.EstateResponse{
				Id: &mockUUID,
			},
			mockError:     nil,
			expectedError: nil,
			prepareMock: func(mockService *service.MockServiceInterface) {
				mockService.EXPECT().PostEstate(gomock.Any(), gomock.Any()).Return(generated.EstateResponse{Id: &mockUUID}, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Invalid JSON",
			requestBody:    `{"width": 5, "length": "invalid"}`,
			mockResponse:   generated.EstateResponse{},
			mockError:      nil,
			expectedError:  ptr("Invalid request"),
			prepareMock:    func(mockService *service.MockServiceInterface) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid Parameters",
			requestBody:    `{"width": -1, "length": 10}`,
			mockResponse:   generated.EstateResponse{},
			mockError:      nil,
			expectedError:  ptr("Key: 'EstateRequest.Width' Error:Field validation for 'Width' failed"),
			prepareMock:    func(mockService *service.MockServiceInterface) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:          "Service Error",
			requestBody:   `{"width": 5, "length": 10}`,
			mockResponse:  generated.EstateResponse{},
			mockError:     nil,
			expectedError: nil,
			prepareMock: func(mockService *service.MockServiceInterface) {
				mockService.EXPECT().PostEstate(gomock.Any(), gomock.Any()).Return(generated.EstateResponse{}, errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Length Exceeds",
			requestBody:    `{"width": 1, "length": 100000}`,
			mockResponse:   generated.EstateResponse{},
			mockError:      nil,
			expectedError:  ptr("Key: 'EstateRequest.Length' Error:Field validation for 'Length' failed"),
			prepareMock:    func(mockService *service.MockServiceInterface) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Width Exceeds",
			requestBody:    `{"width": 1000000, "length": 1}`,
			mockResponse:   generated.EstateResponse{},
			mockError:      nil,
			expectedError:  ptr("Key: 'EstateRequest.Width' Error:Field validation for 'Width' failed"),
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

			err := server.PostEstate(c)

			assert.Equal(t, tc.expectedStatus, rec.Code)

			if tc.expectedStatus == http.StatusOK {
				var resp generated.EstateResponse
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

func ptr(s string) *string {
	return &s
}
