package dto

import "time"

// ExerciseHistoryResponse represents the history of an exercise
type ExerciseHistoryResponse struct {
	ExerciseName string                `json:"exercise_name"`
	Records      []ExerciseRecord      `json:"records"`
	OneRepMax    *OneRepMaxRecord      `json:"one_rep_max,omitempty"`
	TotalVolume  float64               `json:"total_volume"`
	TotalSets    int                   `json:"total_sets"`
}

// ExerciseRecord represents a single exercise record
type ExerciseRecord struct {
	Date         time.Time `json:"date"`
	SessionID    string    `json:"session_id"`
	Weight       float64   `json:"weight"`
	Reps         int       `json:"reps"`
	Volume       float64   `json:"volume"` // weight * reps
	EstimatedORM float64   `json:"estimated_1rm"`
	RPE          *float64  `json:"rpe,omitempty"`
}

// OneRepMaxRecord represents a one rep max record
type OneRepMaxRecord struct {
	Weight float64   `json:"weight"`
	Date   time.Time `json:"date"`
}

// ACWRResponse represents Acute:Chronic Workload Ratio
type ACWRResponse struct {
	ExerciseName   string             `json:"exercise_name,omitempty"`
	CurrentACWR    float64            `json:"current_acwr"`
	AcuteLoad      float64            `json:"acute_load"`      // Last 7 days
	ChronicLoad    float64            `json:"chronic_load"`    // Last 28 days
	Status         string             `json:"status"`          // "optimal", "high_risk", "undertraining"
	Recommendation string             `json:"recommendation"`
	History        []ACWRDataPoint    `json:"history,omitempty"`
}

// ACWRDataPoint represents a single ACWR data point over time
type ACWRDataPoint struct {
	Date        time.Time `json:"date"`
	ACWR        float64   `json:"acwr"`
	AcuteLoad   float64   `json:"acute_load"`
	ChronicLoad float64   `json:"chronic_load"`
}

// VolumeProgressResponse represents volume progress over time
type VolumeProgressResponse struct {
	ExerciseName string              `json:"exercise_name,omitempty"`
	Period       string              `json:"period"` // "week", "month", "year"
	DataPoints   []VolumeDataPoint   `json:"data_points"`
	TotalVolume  float64             `json:"total_volume"`
	AverageVolume float64            `json:"average_volume"`
	Trend        string              `json:"trend"` // "increasing", "stable", "decreasing"
}

// VolumeDataPoint represents volume at a specific time
type VolumeDataPoint struct {
	Date   time.Time `json:"date"`
	Volume float64   `json:"volume"`
	Sets   int       `json:"sets"`
}

// ProgressSummaryResponse represents overall progress summary
type ProgressSummaryResponse struct {
	Period           string                   `json:"period"`
	TotalSessions    int                      `json:"total_sessions"`
	TotalSets        int                      `json:"total_sets"`
	TotalVolume      float64                  `json:"total_volume"`
	AverageACWR      float64                  `json:"average_acwr"`
	TopExercises     []ExerciseVolumeSummary  `json:"top_exercises"`
	OneRepMaxRecords []OneRepMaxSummary       `json:"one_rep_max_records"`
}

// ExerciseVolumeSummary represents volume summary for an exercise
type ExerciseVolumeSummary struct {
	ExerciseName string  `json:"exercise_name"`
	TotalVolume  float64 `json:"total_volume"`
	TotalSets    int     `json:"total_sets"`
	MaxWeight    float64 `json:"max_weight"`
}

// OneRepMaxSummary represents a 1RM summary
type OneRepMaxSummary struct {
	ExerciseName string    `json:"exercise_name"`
	Weight       float64   `json:"weight"`
	Date         time.Time `json:"date"`
}

// AnalyticsQueryParams represents query parameters for analytics
type AnalyticsQueryParams struct {
	ExerciseName string
	StartDate    *time.Time
	EndDate      *time.Time
	Period       string // "week", "month", "year"
}
