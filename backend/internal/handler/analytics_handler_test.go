package handler_test

import (
	"context"
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

// MockAnalyticsService is a mock implementation of AnalyticsServiceInterface
type MockAnalyticsService struct {
	mock.Mock
}

func (m *MockAnalyticsService) GetExerciseHistory(ctx context.Context, userID uuid.UUID, exerciseName string, startDate, endDate time.Time) (*dto.ExerciseHistoryResponse, error) {
	args := m.Called(ctx, userID, exerciseName, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ExerciseHistoryResponse), args.Error(1)
}

func (m *MockAnalyticsService) CalculateACWR(ctx context.Context, userID uuid.UUID, exerciseName string) (*dto.ACWRResponse, error) {
	args := m.Called(ctx, userID, exerciseName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ACWRResponse), args.Error(1)
}

func (m *MockAnalyticsService) GetVolumeProgress(ctx context.Context, userID uuid.UUID, exerciseName string, period string) (*dto.VolumeProgressResponse, error) {
	args := m.Called(ctx, userID, exerciseName, period)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.VolumeProgressResponse), args.Error(1)
}

func (m *MockAnalyticsService) GetProgressSummary(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) (*dto.ProgressSummaryResponse, error) {
	args := m.Called(ctx, userID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ProgressSummaryResponse), args.Error(1)
}

func TestAnalyticsHandler_GetExerciseHistory(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()
	now := time.Now()
	rpe := 8.0

	tests := []struct {
		name           string
		userID         *uuid.UUID
		exerciseName   string
		queryParams    map[string]string
		mockResponse   *dto.ExerciseHistoryResponse
		mockError      error
		expectedStatus int
	}{
		{
			name:         "successful exercise history retrieval",
			userID:       &userID,
			exerciseName: "Squat",
			queryParams:  map[string]string{"start_date": "2024-01-01", "end_date": "2024-12-31"},
			mockResponse: &dto.ExerciseHistoryResponse{
				ExerciseName: "Squat",
				Records: []dto.ExerciseRecord{
					{
						Date:         now,
						SessionID:    uuid.New().String(),
						Weight:       100.0,
						Reps:         5,
						Volume:       500.0,
						EstimatedORM: 112.5,
						RPE:          &rpe,
					},
				},
				OneRepMax: &dto.OneRepMaxRecord{
					Weight: 150.0,
					Date:   now,
				},
				TotalVolume: 500.0,
				TotalSets:   1,
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing user authentication",
			userID:         nil,
			exerciseName:   "Squat",
			queryParams:    map[string]string{},
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "missing exercise name",
			userID:         &userID,
			exerciseName:   "",
			queryParams:    map[string]string{},
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "service error",
			userID:         &userID,
			exerciseName:   "Squat",
			queryParams:    map[string]string{},
			mockResponse:   nil,
			mockError:      fmt.Errorf("database error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock service
			mockService := new(MockAnalyticsService)
			if tt.mockResponse != nil || tt.mockError != nil {
				mockService.On("GetExerciseHistory", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(tt.mockResponse, tt.mockError)
			}

			// Create handler
			analyticsHandler := handler.NewAnalyticsHandler(mockService)

			// Create test context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Set user ID in context if provided
			if tt.userID != nil {
				c.Set("user_id", *tt.userID)
			}

			// Set exercise name param
			c.Params = gin.Params{{Key: "name", Value: tt.exerciseName}}

			// Build query string
			queryString := ""
			for k, v := range tt.queryParams {
				if queryString != "" {
					queryString += "&"
				}
				queryString += fmt.Sprintf("%s=%s", k, v)
			}

			// Prepare request
			url := fmt.Sprintf("/v1/analytics/exercise/%s", tt.exerciseName)
			if queryString != "" {
				url += "?" + queryString
			}
			c.Request = httptest.NewRequest(http.MethodGet, url, nil)

			// Call handler
			analyticsHandler.GetExerciseHistory(c)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.mockResponse != nil || tt.mockError != nil {
				mockService.AssertExpectations(t)
			}
		})
	}
}

func TestAnalyticsHandler_GetACWR(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()

	tests := []struct {
		name           string
		userID         *uuid.UUID
		queryParams    map[string]string
		mockResponse   *dto.ACWRResponse
		mockError      error
		expectedStatus int
	}{
		{
			name:        "successful ACWR calculation",
			userID:      &userID,
			queryParams: map[string]string{"exercise": "Squat"},
			mockResponse: &dto.ACWRResponse{
				ExerciseName:   "Squat",
				CurrentACWR:    1.2,
				AcuteLoad:      2400.0,
				ChronicLoad:    2000.0,
				Status:         "optimal",
				Recommendation: "Training load is in the optimal range for performance and injury prevention",
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
			name:        "all exercises ACWR",
			userID:      &userID,
			queryParams: map[string]string{},
			mockResponse: &dto.ACWRResponse{
				ExerciseName:   "",
				CurrentACWR:    1.1,
				AcuteLoad:      5000.0,
				ChronicLoad:    4545.0,
				Status:         "optimal",
				Recommendation: "Training load is in the optimal range for performance and injury prevention",
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "service error",
			userID:         &userID,
			queryParams:    map[string]string{"exercise": "Squat"},
			mockResponse:   nil,
			mockError:      fmt.Errorf("calculation failed"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock service
			mockService := new(MockAnalyticsService)
			if tt.mockResponse != nil || tt.mockError != nil {
				mockService.On("CalculateACWR", mock.Anything, mock.Anything, mock.Anything).
					Return(tt.mockResponse, tt.mockError)
			}

			// Create handler
			analyticsHandler := handler.NewAnalyticsHandler(mockService)

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
			url := "/v1/analytics/acwr"
			if queryString != "" {
				url += "?" + queryString
			}
			c.Request = httptest.NewRequest(http.MethodGet, url, nil)

			// Call handler
			analyticsHandler.GetACWR(c)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.mockResponse != nil || tt.mockError != nil {
				mockService.AssertExpectations(t)
			}
		})
	}
}

func TestAnalyticsHandler_GetVolumeProgress(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()
	now := time.Now()

	tests := []struct {
		name           string
		userID         *uuid.UUID
		queryParams    map[string]string
		mockResponse   *dto.VolumeProgressResponse
		mockError      error
		expectedStatus int
	}{
		{
			name:        "successful volume progress retrieval",
			userID:      &userID,
			queryParams: map[string]string{"exercise": "Squat", "period": "month"},
			mockResponse: &dto.VolumeProgressResponse{
				ExerciseName: "Squat",
				Period:       "month",
				DataPoints: []dto.VolumeDataPoint{
					{Date: now.AddDate(0, 0, -7), Volume: 2000.0, Sets: 10},
					{Date: now, Volume: 2400.0, Sets: 12},
				},
				TotalVolume:   4400.0,
				AverageVolume: 2200.0,
				Trend:         "increasing",
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
			name:           "invalid period parameter",
			userID:         &userID,
			queryParams:    map[string]string{"period": "invalid"},
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "service error",
			userID:         &userID,
			queryParams:    map[string]string{"period": "month"},
			mockResponse:   nil,
			mockError:      fmt.Errorf("database error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock service
			mockService := new(MockAnalyticsService)
			if tt.mockResponse != nil || tt.mockError != nil {
				mockService.On("GetVolumeProgress", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(tt.mockResponse, tt.mockError)
			}

			// Create handler
			analyticsHandler := handler.NewAnalyticsHandler(mockService)

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
			url := "/v1/analytics/volume"
			if queryString != "" {
				url += "?" + queryString
			}
			c.Request = httptest.NewRequest(http.MethodGet, url, nil)

			// Call handler
			analyticsHandler.GetVolumeProgress(c)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.mockResponse != nil || tt.mockError != nil {
				mockService.AssertExpectations(t)
			}
		})
	}
}

func TestAnalyticsHandler_GetProgressSummary(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()
	now := time.Now()

	tests := []struct {
		name           string
		userID         *uuid.UUID
		queryParams    map[string]string
		mockResponse   *dto.ProgressSummaryResponse
		mockError      error
		expectedStatus int
	}{
		{
			name:        "successful progress summary retrieval",
			userID:      &userID,
			queryParams: map[string]string{"start_date": "2024-01-01", "end_date": "2024-12-31"},
			mockResponse: &dto.ProgressSummaryResponse{
				Period:        "2024-01-01 to 2024-12-31",
				TotalSessions: 50,
				TotalSets:     500,
				TotalVolume:   50000.0,
				TopExercises: []dto.ExerciseVolumeSummary{
					{ExerciseName: "Squat", TotalVolume: 20000.0, TotalSets: 200, MaxWeight: 150.0},
					{ExerciseName: "Bench Press", TotalVolume: 15000.0, TotalSets: 150, MaxWeight: 100.0},
				},
				OneRepMaxRecords: []dto.OneRepMaxSummary{
					{ExerciseName: "Squat", Weight: 150.0, Date: now},
					{ExerciseName: "Bench Press", Weight: 100.0, Date: now},
				},
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
			name:        "default date range",
			userID:      &userID,
			queryParams: map[string]string{},
			mockResponse: &dto.ProgressSummaryResponse{
				Period:        "Last 30 days",
				TotalSessions: 10,
				TotalSets:     100,
				TotalVolume:   10000.0,
				TopExercises: []dto.ExerciseVolumeSummary{
					{ExerciseName: "Squat", TotalVolume: 6000.0, TotalSets: 60, MaxWeight: 150.0},
				},
				OneRepMaxRecords: []dto.OneRepMaxSummary{
					{ExerciseName: "Squat", Weight: 150.0, Date: now},
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "service error",
			userID:         &userID,
			queryParams:    map[string]string{},
			mockResponse:   nil,
			mockError:      fmt.Errorf("database error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock service
			mockService := new(MockAnalyticsService)
			if tt.mockResponse != nil || tt.mockError != nil {
				mockService.On("GetProgressSummary", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(tt.mockResponse, tt.mockError)
			}

			// Create handler
			analyticsHandler := handler.NewAnalyticsHandler(mockService)

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
			url := "/v1/analytics/summary"
			if queryString != "" {
				url += "?" + queryString
			}
			c.Request = httptest.NewRequest(http.MethodGet, url, nil)

			// Call handler
			analyticsHandler.GetProgressSummary(c)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.mockResponse != nil || tt.mockError != nil {
				mockService.AssertExpectations(t)
			}
		})
	}
}
