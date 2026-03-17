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

// MockSessionService is a mock implementation of SessionServiceInterface
type MockSessionService struct {
	mock.Mock
}

func (m *MockSessionService) CreateSession(ctx context.Context, userID uuid.UUID, req *dto.CreateSessionRequest) (*dto.SessionResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.SessionResponse), args.Error(1)
}

func (m *MockSessionService) GetSession(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID) (*dto.SessionResponse, error) {
	args := m.Called(ctx, userID, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.SessionResponse), args.Error(1)
}

func (m *MockSessionService) ListSessions(ctx context.Context, userID uuid.UUID, page, pageSize int) (*dto.SessionListResponse, error) {
	args := m.Called(ctx, userID, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.SessionListResponse), args.Error(1)
}

func (m *MockSessionService) UpdateSession(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID, req *dto.UpdateSessionRequest) (*dto.SessionResponse, error) {
	args := m.Called(ctx, userID, sessionID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.SessionResponse), args.Error(1)
}

func (m *MockSessionService) DeleteSession(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID) error {
	args := m.Called(ctx, userID, sessionID)
	return args.Error(0)
}

func TestSessionHandler_CreateSession(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()
	sessionID := uuid.New()
	now := time.Now()

	tests := []struct {
		name           string
		userID         *uuid.UUID
		requestBody    interface{}
		mockResponse   *dto.SessionResponse
		mockError      error
		expectedStatus int
	}{
		{
			name:   "successful session creation",
			userID: &userID,
			requestBody: dto.CreateSessionRequest{
				Date:  now,
				Notes: stringPtr("Great workout"),
				Sets: []dto.CreateSetInput{
					{
						ExerciseName: "Squat",
						Weight:       100.0,
						Reps:         5,
						RPE:          float64Ptr(8.0),
						SetOrder:     1,
					},
				},
			},
			mockResponse: &dto.SessionResponse{
				ID:        sessionID,
				UserID:    userID,
				Date:      now,
				Notes:     stringPtr("Great workout"),
				Sets:      []dto.SetResponse{{ExerciseName: "Squat", Weight: 100.0, Reps: 5, RPE: float64Ptr(8.0), SetOrder: 1}},
				CreatedAt: now,
				UpdatedAt: now,
			},
			mockError:      nil,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "missing user authentication",
			userID:         nil,
			requestBody:    dto.CreateSessionRequest{Date: now},
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:   "missing required date field",
			userID: &userID,
			requestBody: map[string]interface{}{
				"notes": "Missing date",
			},
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "service error",
			userID: &userID,
			requestBody: dto.CreateSessionRequest{
				Date: now,
			},
			mockResponse:   nil,
			mockError:      fmt.Errorf("database error"),
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock service
			mockService := new(MockSessionService)
			if tt.mockResponse != nil || tt.mockError != nil {
				mockService.On("CreateSession", mock.Anything, mock.Anything, mock.Anything).
					Return(tt.mockResponse, tt.mockError)
			}

			// Create handler
			sessionHandler := handler.NewSessionHandler(mockService)

			// Create test context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Set user ID in context if provided
			if tt.userID != nil {
				c.Set("user_id", *tt.userID)
			}

			// Prepare request
			bodyBytes, _ := json.Marshal(tt.requestBody)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/sessions", bytes.NewBuffer(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")

			// Call handler
			sessionHandler.CreateSession(c)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.mockResponse != nil || tt.mockError != nil {
				mockService.AssertExpectations(t)
			}
		})
	}
}

func TestSessionHandler_GetSession(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()
	sessionID := uuid.New()
	now := time.Now()

	tests := []struct {
		name           string
		userID         *uuid.UUID
		sessionID      string
		mockResponse   *dto.SessionResponse
		mockError      error
		expectedStatus int
	}{
		{
			name:      "successful session retrieval",
			userID:    &userID,
			sessionID: sessionID.String(),
			mockResponse: &dto.SessionResponse{
				ID:        sessionID,
				UserID:    userID,
				Date:      now,
				Notes:     stringPtr("Test session"),
				Sets:      []dto.SetResponse{},
				CreatedAt: now,
				UpdatedAt: now,
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing user authentication",
			userID:         nil,
			sessionID:      sessionID.String(),
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid session ID",
			userID:         &userID,
			sessionID:      "invalid-uuid",
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "session not found",
			userID:         &userID,
			sessionID:      sessionID.String(),
			mockResponse:   nil,
			mockError:      fmt.Errorf("session not found"),
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock service
			mockService := new(MockSessionService)
			if tt.mockResponse != nil || tt.mockError != nil {
				mockService.On("GetSession", mock.Anything, mock.Anything, mock.Anything).
					Return(tt.mockResponse, tt.mockError)
			}

			// Create handler
			sessionHandler := handler.NewSessionHandler(mockService)

			// Create test context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Set user ID in context if provided
			if tt.userID != nil {
				c.Set("user_id", *tt.userID)
			}

			// Set session ID param
			c.Params = gin.Params{{Key: "id", Value: tt.sessionID}}

			// Prepare request
			c.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/sessions/%s", tt.sessionID), nil)

			// Call handler
			sessionHandler.GetSession(c)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.mockResponse != nil || tt.mockError != nil {
				mockService.AssertExpectations(t)
			}
		})
	}
}

func TestSessionHandler_ListSessions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()
	now := time.Now()

	tests := []struct {
		name           string
		userID         *uuid.UUID
		queryParams    map[string]string
		mockResponse   *dto.SessionListResponse
		mockError      error
		expectedStatus int
	}{
		{
			name:        "successful session list",
			userID:      &userID,
			queryParams: map[string]string{"page": "1", "page_size": "10"},
			mockResponse: &dto.SessionListResponse{
				Sessions: []dto.SessionResponse{
					{ID: uuid.New(), UserID: userID, Date: now, Notes: stringPtr("Session 1")},
					{ID: uuid.New(), UserID: userID, Date: now, Notes: stringPtr("Session 2")},
				},
				Total:    2,
				Page:     1,
				PageSize: 10,
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
			mockResponse: &dto.SessionListResponse{
				Sessions: []dto.SessionResponse{},
				Total:    0,
				Page:     1,
				PageSize: 20,
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:        "invalid page parameter defaults to page 1",
			userID:      &userID,
			queryParams: map[string]string{"page": "invalid"},
			mockResponse: &dto.SessionListResponse{
				Sessions: []dto.SessionResponse{},
				Total:    0,
				Page:     1,
				PageSize: 20,
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock service
			mockService := new(MockSessionService)
			if tt.mockResponse != nil || tt.mockError != nil {
				mockService.On("ListSessions", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(tt.mockResponse, tt.mockError)
			}

			// Create handler
			sessionHandler := handler.NewSessionHandler(mockService)

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
			url := "/v1/sessions"
			if queryString != "" {
				url += "?" + queryString
			}
			c.Request = httptest.NewRequest(http.MethodGet, url, nil)

			// Call handler
			sessionHandler.ListSessions(c)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.mockResponse != nil || tt.mockError != nil {
				mockService.AssertExpectations(t)
			}
		})
	}
}

func TestSessionHandler_UpdateSession(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()
	sessionID := uuid.New()
	now := time.Now()

	tests := []struct {
		name           string
		userID         *uuid.UUID
		sessionID      string
		requestBody    interface{}
		mockResponse   *dto.SessionResponse
		mockError      error
		expectedStatus int
	}{
		{
			name:      "successful session update",
			userID:    &userID,
			sessionID: sessionID.String(),
			requestBody: dto.UpdateSessionRequest{
				Notes: stringPtr("Updated notes"),
			},
			mockResponse: &dto.SessionResponse{
				ID:        sessionID,
				UserID:    userID,
				Date:      now,
				Notes:     stringPtr("Updated notes"),
				Sets:      []dto.SetResponse{},
				CreatedAt: now,
				UpdatedAt: now,
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing user authentication",
			userID:         nil,
			sessionID:      sessionID.String(),
			requestBody:    dto.UpdateSessionRequest{},
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid session ID",
			userID:         &userID,
			sessionID:      "invalid-uuid",
			requestBody:    dto.UpdateSessionRequest{},
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "session not found",
			userID:    &userID,
			sessionID: sessionID.String(),
			requestBody: dto.UpdateSessionRequest{
				Notes: stringPtr("Updated"),
			},
			mockResponse:   nil,
			mockError:      fmt.Errorf("session not found"),
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock service
			mockService := new(MockSessionService)
			if tt.mockResponse != nil || tt.mockError != nil {
				mockService.On("UpdateSession", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(tt.mockResponse, tt.mockError)
			}

			// Create handler
			sessionHandler := handler.NewSessionHandler(mockService)

			// Create test context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Set user ID in context if provided
			if tt.userID != nil {
				c.Set("user_id", *tt.userID)
			}

			// Set session ID param
			c.Params = gin.Params{{Key: "id", Value: tt.sessionID}}

			// Prepare request
			bodyBytes, _ := json.Marshal(tt.requestBody)
			c.Request = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/v1/sessions/%s", tt.sessionID), bytes.NewBuffer(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")

			// Call handler
			sessionHandler.UpdateSession(c)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.mockResponse != nil || tt.mockError != nil {
				mockService.AssertExpectations(t)
			}
		})
	}
}

func TestSessionHandler_DeleteSession(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()
	sessionID := uuid.New()

	tests := []struct {
		name           string
		userID         *uuid.UUID
		sessionID      string
		mockError      error
		expectedStatus int
	}{
		{
			name:           "successful session deletion",
			userID:         &userID,
			sessionID:      sessionID.String(),
			mockError:      nil,
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "missing user authentication",
			userID:         nil,
			sessionID:      sessionID.String(),
			mockError:      nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid session ID",
			userID:         &userID,
			sessionID:      "invalid-uuid",
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "session not found",
			userID:         &userID,
			sessionID:      sessionID.String(),
			mockError:      fmt.Errorf("session not found"),
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock service
			mockService := new(MockSessionService)
			if tt.userID != nil && tt.sessionID != "invalid-uuid" {
				mockService.On("DeleteSession", mock.Anything, mock.Anything, mock.Anything).
					Return(tt.mockError)
			}

			// Create handler
			sessionHandler := handler.NewSessionHandler(mockService)

			// Create test context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Set user ID in context if provided
			if tt.userID != nil {
				c.Set("user_id", *tt.userID)
			}

			// Set session ID param
			c.Params = gin.Params{{Key: "id", Value: tt.sessionID}}

			// Prepare request
			c.Request = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/v1/sessions/%s", tt.sessionID), nil)

			// Call handler
			sessionHandler.DeleteSession(c)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.userID != nil && tt.sessionID != "invalid-uuid" {
				mockService.AssertExpectations(t)
			}
		})
	}
}

// Helper functions for pointers
func stringPtr(s string) *string {
	return &s
}

func float64Ptr(f float64) *float64 {
	return &f
}
