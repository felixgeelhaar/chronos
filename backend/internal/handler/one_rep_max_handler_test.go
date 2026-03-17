package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ascend/api/internal/dto"
	"github.com/ascend/api/internal/handler"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockOneRepMaxService is a mock implementation of OneRepMaxServiceInterface
type MockOneRepMaxService struct {
	mock.Mock
}

func (m *MockOneRepMaxService) CreateOneRepMax(ctx context.Context, userID uuid.UUID, req *dto.CreateOneRepMaxRequest) (*dto.OneRepMaxResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.OneRepMaxResponse), args.Error(1)
}

func (m *MockOneRepMaxService) GetOneRepMax(ctx context.Context, userID uuid.UUID, id uuid.UUID) (*dto.OneRepMaxResponse, error) {
	args := m.Called(ctx, userID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.OneRepMaxResponse), args.Error(1)
}

func (m *MockOneRepMaxService) ListOneRepMaxes(ctx context.Context, userID uuid.UUID, page, pageSize int) (*dto.OneRepMaxListResponse, error) {
	args := m.Called(ctx, userID, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.OneRepMaxListResponse), args.Error(1)
}

func (m *MockOneRepMaxService) GetOneRepMaxHistory(ctx context.Context, userID uuid.UUID, exerciseName string) (*dto.OneRepMaxHistoryResponse, error) {
	args := m.Called(ctx, userID, exerciseName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.OneRepMaxHistoryResponse), args.Error(1)
}

func (m *MockOneRepMaxService) UpdateOneRepMax(ctx context.Context, userID uuid.UUID, id uuid.UUID, req *dto.UpdateOneRepMaxRequest) (*dto.OneRepMaxResponse, error) {
	args := m.Called(ctx, userID, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.OneRepMaxResponse), args.Error(1)
}

func (m *MockOneRepMaxService) DeleteOneRepMax(ctx context.Context, userID uuid.UUID, id uuid.UUID) error {
	args := m.Called(ctx, userID, id)
	return args.Error(0)
}

func TestOneRepMaxHandler_CreateOneRepMax(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()
	ormID := uuid.New()
	now := time.Now()

	tests := []struct {
		name           string
		userID         *uuid.UUID
		requestBody    interface{}
		mockResponse   *dto.OneRepMaxResponse
		mockError      error
		expectedStatus int
	}{
		{
			name:   "successful 1RM creation",
			userID: &userID,
			requestBody: dto.CreateOneRepMaxRequest{
				ExerciseName: "Squat",
				Weight:       150.0,
				Date:         now,
			},
			mockResponse: &dto.OneRepMaxResponse{
				ID:           ormID,
				UserID:       userID,
				ExerciseName: "Squat",
				Weight:       150.0,
				Date:         now,
				CreatedAt:    now,
				UpdatedAt:    now,
			},
			mockError:      nil,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "missing user authentication",
			userID:         nil,
			requestBody:    dto.CreateOneRepMaxRequest{ExerciseName: "Squat", Weight: 150.0, Date: now},
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:   "missing required exercise name",
			userID: &userID,
			requestBody: map[string]interface{}{
				"weight": 150.0,
				"date":   now,
			},
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "service error",
			userID: &userID,
			requestBody: dto.CreateOneRepMaxRequest{
				ExerciseName: "Squat",
				Weight:       150.0,
				Date:         now,
			},
			mockResponse:   nil,
			mockError:      fmt.Errorf("database error"),
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock service
			mockService := new(MockOneRepMaxService)
			if tt.mockResponse != nil || tt.mockError != nil {
				mockService.On("CreateOneRepMax", mock.Anything, mock.Anything, mock.Anything).
					Return(tt.mockResponse, tt.mockError)
			}

			// Create handler
			ormHandler := handler.NewOneRepMaxHandler(mockService)

			// Create test context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Set user ID in context if provided
			if tt.userID != nil {
				c.Set("user_id", *tt.userID)
			}

			// Prepare request
			bodyBytes, _ := json.Marshal(tt.requestBody)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/one-rep-maxes", bytes.NewBuffer(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")

			// Call handler
			ormHandler.CreateOneRepMax(c)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.mockResponse != nil || tt.mockError != nil {
				mockService.AssertExpectations(t)
			}
		})
	}
}

func TestOneRepMaxHandler_GetOneRepMax(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()
	ormID := uuid.New()
	now := time.Now()

	tests := []struct {
		name           string
		userID         *uuid.UUID
		ormID          string
		mockResponse   *dto.OneRepMaxResponse
		mockError      error
		expectedStatus int
	}{
		{
			name:   "successful 1RM retrieval",
			userID: &userID,
			ormID:  ormID.String(),
			mockResponse: &dto.OneRepMaxResponse{
				ID:           ormID,
				UserID:       userID,
				ExerciseName: "Squat",
				Weight:       150.0,
				Date:         now,
				CreatedAt:    now,
				UpdatedAt:    now,
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing user authentication",
			userID:         nil,
			ormID:          ormID.String(),
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid 1RM ID",
			userID:         &userID,
			ormID:          "invalid-uuid",
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "1RM not found",
			userID:         &userID,
			ormID:          ormID.String(),
			mockResponse:   nil,
			mockError:      fmt.Errorf("not found"),
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock service
			mockService := new(MockOneRepMaxService)
			if tt.mockResponse != nil || tt.mockError != nil {
				mockService.On("GetOneRepMax", mock.Anything, mock.Anything, mock.Anything).
					Return(tt.mockResponse, tt.mockError)
			}

			// Create handler
			ormHandler := handler.NewOneRepMaxHandler(mockService)

			// Create test context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Set user ID in context if provided
			if tt.userID != nil {
				c.Set("user_id", *tt.userID)
			}

			// Set ID param
			c.Params = gin.Params{{Key: "id", Value: tt.ormID}}

			// Prepare request
			c.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/one-rep-maxes/%s", tt.ormID), nil)

			// Call handler
			ormHandler.GetOneRepMax(c)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.mockResponse != nil || tt.mockError != nil {
				mockService.AssertExpectations(t)
			}
		})
	}
}

func TestOneRepMaxHandler_ListOneRepMaxes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()
	now := time.Now()

	tests := []struct {
		name           string
		userID         *uuid.UUID
		queryParams    map[string]string
		mockResponse   *dto.OneRepMaxListResponse
		mockError      error
		expectedStatus int
	}{
		{
			name:        "successful 1RM list",
			userID:      &userID,
			queryParams: map[string]string{"page": "1", "page_size": "10"},
			mockResponse: &dto.OneRepMaxListResponse{
				Records: []dto.OneRepMaxResponse{
					{ID: uuid.New(), UserID: userID, ExerciseName: "Squat", Weight: 150.0, Date: now},
					{ID: uuid.New(), UserID: userID, ExerciseName: "Bench Press", Weight: 100.0, Date: now},
				},
				Total:      2,
				Page:       1,
				PageSize:   10,
				TotalPages: 1,
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing user authentication",
			userID:         nil,
			queryParams:    map[string]string{},
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:        "default pagination",
			userID:      &userID,
			queryParams: map[string]string{},
			mockResponse: &dto.OneRepMaxListResponse{
				Records:    []dto.OneRepMaxResponse{},
				Total:      0,
				Page:       1,
				PageSize:   20,
				TotalPages: 0,
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:        "service error",
			userID:      &userID,
			queryParams: map[string]string{"page": "1"},
			mockResponse: nil,
			mockError:   fmt.Errorf("database error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock service
			mockService := new(MockOneRepMaxService)
			if tt.mockResponse != nil || tt.mockError != nil {
				mockService.On("ListOneRepMaxes", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(tt.mockResponse, tt.mockError)
			}

			// Create handler
			ormHandler := handler.NewOneRepMaxHandler(mockService)

			// Create test context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Set user ID in context if provided
			if tt.userID != nil {
				c.Set("user_id", *tt.userID)
			}

			// Build query string
			queryString := ""
			for k, v := range tt.queryParams {
				if queryString != "" {
					queryString += "&"
				}
				queryString += fmt.Sprintf("%s=%s", k, v)
			}

			// Prepare request
			url := "/v1/one-rep-maxes"
			if queryString != "" {
				url += "?" + queryString
			}
			c.Request = httptest.NewRequest(http.MethodGet, url, nil)

			// Call handler
			ormHandler.ListOneRepMaxes(c)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.mockResponse != nil || tt.mockError != nil {
				mockService.AssertExpectations(t)
			}
		})
	}
}

func TestOneRepMaxHandler_GetOneRepMaxHistory(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()
	now := time.Now()

	tests := []struct {
		name           string
		userID         *uuid.UUID
		exerciseName   string
		mockResponse   *dto.OneRepMaxHistoryResponse
		mockError      error
		expectedStatus int
	}{
		{
			name:         "successful history retrieval",
			userID:       &userID,
			exerciseName: "Squat",
			mockResponse: &dto.OneRepMaxHistoryResponse{
				ExerciseName: "Squat",
				CurrentRecord: &dto.OneRepMaxResponse{
					ID:           uuid.New(),
					UserID:       userID,
					ExerciseName: "Squat",
					Weight:       150.0,
					Date:         now,
				},
				History: []dto.OneRepMaxResponse{
					{ID: uuid.New(), UserID: userID, ExerciseName: "Squat", Weight: 140.0, Date: now.AddDate(0, -1, 0)},
					{ID: uuid.New(), UserID: userID, ExerciseName: "Squat", Weight: 150.0, Date: now},
				},
				PersonalBest: 150.0,
				Improvement:  7.14,
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing user authentication",
			userID:         nil,
			exerciseName:   "Squat",
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "missing exercise name",
			userID:         &userID,
			exerciseName:   "",
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "service error",
			userID:         &userID,
			exerciseName:   "Squat",
			mockResponse:   nil,
			mockError:      fmt.Errorf("database error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock service
			mockService := new(MockOneRepMaxService)
			if tt.mockResponse != nil || tt.mockError != nil {
				mockService.On("GetOneRepMaxHistory", mock.Anything, mock.Anything, mock.Anything).
					Return(tt.mockResponse, tt.mockError)
			}

			// Create handler
			ormHandler := handler.NewOneRepMaxHandler(mockService)

			// Create test context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Set user ID in context if provided
			if tt.userID != nil {
				c.Set("user_id", *tt.userID)
			}

			// Set exercise name param
			c.Params = gin.Params{{Key: "name", Value: tt.exerciseName}}

			// Prepare request
			c.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/one-rep-maxes/exercise/%s/history", tt.exerciseName), nil)

			// Call handler
			ormHandler.GetOneRepMaxHistory(c)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.mockResponse != nil || tt.mockError != nil {
				mockService.AssertExpectations(t)
			}
		})
	}
}

func TestOneRepMaxHandler_UpdateOneRepMax(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()
	ormID := uuid.New()
	now := time.Now()
	updatedWeight := 160.0

	tests := []struct {
		name           string
		userID         *uuid.UUID
		ormID          string
		requestBody    interface{}
		mockResponse   *dto.OneRepMaxResponse
		mockError      error
		expectedStatus int
	}{
		{
			name:   "successful 1RM update",
			userID: &userID,
			ormID:  ormID.String(),
			requestBody: dto.UpdateOneRepMaxRequest{
				Weight: &updatedWeight,
			},
			mockResponse: &dto.OneRepMaxResponse{
				ID:           ormID,
				UserID:       userID,
				ExerciseName: "Squat",
				Weight:       updatedWeight,
				Date:         now,
				CreatedAt:    now,
				UpdatedAt:    now,
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing user authentication",
			userID:         nil,
			ormID:          ormID.String(),
			requestBody:    dto.UpdateOneRepMaxRequest{},
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid 1RM ID",
			userID:         &userID,
			ormID:          "invalid-uuid",
			requestBody:    dto.UpdateOneRepMaxRequest{},
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "1RM not found",
			userID: &userID,
			ormID:  ormID.String(),
			requestBody: dto.UpdateOneRepMaxRequest{
				Weight: &updatedWeight,
			},
			mockResponse:   nil,
			mockError:      fmt.Errorf("not found"),
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock service
			mockService := new(MockOneRepMaxService)
			if tt.mockResponse != nil || tt.mockError != nil {
				mockService.On("UpdateOneRepMax", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(tt.mockResponse, tt.mockError)
			}

			// Create handler
			ormHandler := handler.NewOneRepMaxHandler(mockService)

			// Create test context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Set user ID in context if provided
			if tt.userID != nil {
				c.Set("user_id", *tt.userID)
			}

			// Set ID param
			c.Params = gin.Params{{Key: "id", Value: tt.ormID}}

			// Prepare request
			bodyBytes, _ := json.Marshal(tt.requestBody)
			c.Request = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/v1/one-rep-maxes/%s", tt.ormID), bytes.NewBuffer(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")

			// Call handler
			ormHandler.UpdateOneRepMax(c)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.mockResponse != nil || tt.mockError != nil {
				mockService.AssertExpectations(t)
			}
		})
	}
}

func TestOneRepMaxHandler_DeleteOneRepMax(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()
	ormID := uuid.New()

	tests := []struct {
		name           string
		userID         *uuid.UUID
		ormID          string
		mockError      error
		expectedStatus int
	}{
		{
			name:           "successful 1RM deletion",
			userID:         &userID,
			ormID:          ormID.String(),
			mockError:      nil,
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "missing user authentication",
			userID:         nil,
			ormID:          ormID.String(),
			mockError:      nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid 1RM ID",
			userID:         &userID,
			ormID:          "invalid-uuid",
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "1RM not found",
			userID:         &userID,
			ormID:          ormID.String(),
			mockError:      fmt.Errorf("not found"),
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock service
			mockService := new(MockOneRepMaxService)
			if tt.userID != nil && tt.ormID != "invalid-uuid" {
				mockService.On("DeleteOneRepMax", mock.Anything, mock.Anything, mock.Anything).
					Return(tt.mockError)
			}

			// Create handler
			ormHandler := handler.NewOneRepMaxHandler(mockService)

			// Create test context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Set user ID in context if provided
			if tt.userID != nil {
				c.Set("user_id", *tt.userID)
			}

			// Set ID param
			c.Params = gin.Params{{Key: "id", Value: tt.ormID}}

			// Prepare request
			c.Request = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/v1/one-rep-maxes/%s", tt.ormID), nil)

			// Call handler
			ormHandler.DeleteOneRepMax(c)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.userID != nil && tt.ormID != "invalid-uuid" {
				mockService.AssertExpectations(t)
			}
		})
	}
}
