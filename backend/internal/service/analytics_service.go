package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/ascend/api/internal/dto"
	"github.com/ascend/api/internal/repository"
	"github.com/google/uuid"
)

// AnalyticsServiceInterface defines the contract for analytics operations
type AnalyticsServiceInterface interface {
	GetExerciseHistory(ctx context.Context, userID uuid.UUID, exerciseName string, startDate, endDate time.Time) (*dto.ExerciseHistoryResponse, error)
	CalculateACWR(ctx context.Context, userID uuid.UUID, exerciseName string) (*dto.ACWRResponse, error)
	GetVolumeProgress(ctx context.Context, userID uuid.UUID, exerciseName string, period string) (*dto.VolumeProgressResponse, error)
	GetProgressSummary(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) (*dto.ProgressSummaryResponse, error)
}

// Verify that AnalyticsService implements AnalyticsServiceInterface
var _ AnalyticsServiceInterface = (*AnalyticsService)(nil)

// AnalyticsService handles analytics and progress tracking
type AnalyticsService struct {
	sessionRepo    repository.SessionRepository
	setRepo        repository.SetRepository
	oneRepMaxRepo  repository.OneRepMaxRepository
}

// NewAnalyticsService creates a new analytics service
func NewAnalyticsService(
	sessionRepo repository.SessionRepository,
	setRepo repository.SetRepository,
	oneRepMaxRepo repository.OneRepMaxRepository,
) *AnalyticsService {
	return &AnalyticsService{
		sessionRepo:   sessionRepo,
		setRepo:       setRepo,
		oneRepMaxRepo: oneRepMaxRepo,
	}
}

// GetExerciseHistory retrieves the history of a specific exercise
func (s *AnalyticsService) GetExerciseHistory(ctx context.Context, userID uuid.UUID, exerciseName string, startDate, endDate time.Time) (*dto.ExerciseHistoryResponse, error) {
	// Get sessions in date range
	sessions, err := s.sessionRepo.GetByUserIDAndDateRange(ctx, userID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions: %w", err)
	}

	records := []dto.ExerciseRecord{}
	totalVolume := 0.0
	totalSets := 0

	for _, session := range sessions {
		for _, set := range session.Sets {
			if set.ExerciseName == exerciseName {
				volume := set.Weight * float64(set.Reps)
				estimatedORM := calculateEstimated1RM(set.Weight, set.Reps)

				records = append(records, dto.ExerciseRecord{
					Date:         session.Date,
					SessionID:    session.ID.String(),
					Weight:       set.Weight,
					Reps:         set.Reps,
					Volume:       volume,
					EstimatedORM: estimatedORM,
					RPE:          set.RPE,
				})

				totalVolume += volume
				totalSets++
			}
		}
	}

	// Get one rep max
	var oneRepMax *dto.OneRepMaxRecord
	orm, err := s.oneRepMaxRepo.GetLatestByUserIDAndExercise(ctx, userID, exerciseName)
	if err == nil && orm != nil {
		oneRepMax = &dto.OneRepMaxRecord{
			Weight: orm.Weight,
			Date:   orm.Date,
		}
	}

	return &dto.ExerciseHistoryResponse{
		ExerciseName: exerciseName,
		Records:      records,
		OneRepMax:    oneRepMax,
		TotalVolume:  totalVolume,
		TotalSets:    totalSets,
	}, nil
}

// CalculateACWR calculates Acute:Chronic Workload Ratio for an exercise
func (s *AnalyticsService) CalculateACWR(ctx context.Context, userID uuid.UUID, exerciseName string) (*dto.ACWRResponse, error) {
	now := time.Now()

	// Get last 28 days of data (chronic period)
	chronicStart := now.AddDate(0, 0, -28)
	sessions, err := s.sessionRepo.GetByUserIDAndDateRange(ctx, userID, chronicStart, now)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions: %w", err)
	}

	// Calculate daily volumes
	dailyVolumes := make(map[string]float64)
	for _, session := range sessions {
		dateKey := session.Date.Format("2006-01-02")
		for _, set := range session.Sets {
			if exerciseName == "" || set.ExerciseName == exerciseName {
				volume := set.Weight * float64(set.Reps)
				dailyVolumes[dateKey] += volume
			}
		}
	}

	// Calculate acute load (last 7 days average)
	acuteStart := now.AddDate(0, 0, -7)
	acuteLoad := 0.0
	acuteDays := 0
	for i := 0; i < 7; i++ {
		date := acuteStart.AddDate(0, 0, i)
		dateKey := date.Format("2006-01-02")
		if volume, exists := dailyVolumes[dateKey]; exists {
			acuteLoad += volume
		}
		acuteDays++
	}
	acuteLoad = acuteLoad / float64(acuteDays)

	// Calculate chronic load (last 28 days average)
	chronicLoad := 0.0
	chronicDays := 0
	for i := 0; i < 28; i++ {
		date := chronicStart.AddDate(0, 0, i)
		dateKey := date.Format("2006-01-02")
		if volume, exists := dailyVolumes[dateKey]; exists {
			chronicLoad += volume
		}
		chronicDays++
	}
	chronicLoad = chronicLoad / float64(chronicDays)

	// Calculate ACWR
	currentACWR := 0.0
	if chronicLoad > 0 {
		currentACWR = acuteLoad / chronicLoad
	}

	// Determine status and recommendation
	status, recommendation := getACWRStatus(currentACWR)

	return &dto.ACWRResponse{
		ExerciseName:   exerciseName,
		CurrentACWR:    currentACWR,
		AcuteLoad:      acuteLoad,
		ChronicLoad:    chronicLoad,
		Status:         status,
		Recommendation: recommendation,
	}, nil
}

// GetVolumeProgress retrieves volume progress over time
func (s *AnalyticsService) GetVolumeProgress(ctx context.Context, userID uuid.UUID, exerciseName string, period string) (*dto.VolumeProgressResponse, error) {
	// Determine date range based on period
	now := time.Now()
	var startDate time.Time

	switch period {
	case "week":
		startDate = now.AddDate(0, 0, -7)
	case "month":
		startDate = now.AddDate(0, -1, 0)
	case "year":
		startDate = now.AddDate(-1, 0, 0)
	default:
		startDate = now.AddDate(0, -1, 0) // Default to month
		period = "month"
	}

	sessions, err := s.sessionRepo.GetByUserIDAndDateRange(ctx, userID, startDate, now)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions: %w", err)
	}

	// Group by date
	dailyData := make(map[string]*dto.VolumeDataPoint)
	totalVolume := 0.0

	for _, session := range sessions {
		dateKey := session.Date.Format("2006-01-02")
		if _, exists := dailyData[dateKey]; !exists {
			dailyData[dateKey] = &dto.VolumeDataPoint{
				Date:   session.Date,
				Volume: 0,
				Sets:   0,
			}
		}

		for _, set := range session.Sets {
			if exerciseName == "" || set.ExerciseName == exerciseName {
				volume := set.Weight * float64(set.Reps)
				dailyData[dateKey].Volume += volume
				dailyData[dateKey].Sets++
				totalVolume += volume
			}
		}
	}

	// Convert to slice and sort
	dataPoints := make([]dto.VolumeDataPoint, 0, len(dailyData))
	for _, point := range dailyData {
		dataPoints = append(dataPoints, *point)
	}

	// Calculate average
	averageVolume := 0.0
	if len(dataPoints) > 0 {
		averageVolume = totalVolume / float64(len(dataPoints))
	}

	// Determine trend
	trend := calculateTrend(dataPoints)

	return &dto.VolumeProgressResponse{
		ExerciseName:  exerciseName,
		Period:        period,
		DataPoints:    dataPoints,
		TotalVolume:   totalVolume,
		AverageVolume: averageVolume,
		Trend:         trend,
	}, nil
}

// GetProgressSummary retrieves an overall progress summary
func (s *AnalyticsService) GetProgressSummary(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) (*dto.ProgressSummaryResponse, error) {
	sessions, err := s.sessionRepo.GetByUserIDAndDateRange(ctx, userID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions: %w", err)
	}

	totalSessions := len(sessions)
	totalSets := 0
	totalVolume := 0.0
	exerciseVolumes := make(map[string]*dto.ExerciseVolumeSummary)

	for _, session := range sessions {
		for _, set := range session.Sets {
			totalSets++
			volume := set.Weight * float64(set.Reps)
			totalVolume += volume

			if _, exists := exerciseVolumes[set.ExerciseName]; !exists {
				exerciseVolumes[set.ExerciseName] = &dto.ExerciseVolumeSummary{
					ExerciseName: set.ExerciseName,
					TotalVolume:  0,
					TotalSets:    0,
					MaxWeight:    0,
				}
			}

			exerciseVolumes[set.ExerciseName].TotalVolume += volume
			exerciseVolumes[set.ExerciseName].TotalSets++
			if set.Weight > exerciseVolumes[set.ExerciseName].MaxWeight {
				exerciseVolumes[set.ExerciseName].MaxWeight = set.Weight
			}
		}
	}

	// Convert to slice and get top exercises
	topExercises := make([]dto.ExerciseVolumeSummary, 0, len(exerciseVolumes))
	for _, summary := range exerciseVolumes {
		topExercises = append(topExercises, *summary)
	}

	// Get one rep maxes
	orms, _ := s.oneRepMaxRepo.GetByUserID(ctx, userID)
	oneRepMaxRecords := make([]dto.OneRepMaxSummary, 0, len(orms))
	for _, orm := range orms {
		oneRepMaxRecords = append(oneRepMaxRecords, dto.OneRepMaxSummary{
			ExerciseName: orm.ExerciseName,
			Weight:       orm.Weight,
			Date:         orm.Date,
		})
	}

	return &dto.ProgressSummaryResponse{
		Period:           fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		TotalSessions:    totalSessions,
		TotalSets:        totalSets,
		TotalVolume:      totalVolume,
		TopExercises:     topExercises,
		OneRepMaxRecords: oneRepMaxRecords,
	}, nil
}

// calculateEstimated1RM estimates one rep max using Epley formula
func calculateEstimated1RM(weight float64, reps int) float64 {
	if reps == 1 {
		return weight
	}
	// Epley formula: 1RM = weight × (1 + reps/30)
	return weight * (1 + float64(reps)/30.0)
}

// getACWRStatus determines status and recommendation based on ACWR value
func getACWRStatus(acwr float64) (string, string) {
	if acwr < 0.8 {
		return "undertraining", "Consider increasing training volume gradually to improve fitness adaptations"
	} else if acwr >= 0.8 && acwr <= 1.3 {
		return "optimal", "Training load is in the optimal range for performance and injury prevention"
	} else {
		return "high_risk", "Training load may be too high. Consider reducing volume to prevent overtraining"
	}
}

// calculateTrend determines if volume is increasing, stable, or decreasing
func calculateTrend(dataPoints []dto.VolumeDataPoint) string {
	if len(dataPoints) < 2 {
		return "stable"
	}

	// Simple linear regression to determine trend
	n := float64(len(dataPoints))
	var sumX, sumY, sumXY, sumX2 float64

	for i, point := range dataPoints {
		x := float64(i)
		y := point.Volume
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	// Calculate slope
	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)

	// Determine trend based on slope
	if math.Abs(slope) < 10 {
		return "stable"
	} else if slope > 0 {
		return "increasing"
	} else {
		return "decreasing"
	}
}
