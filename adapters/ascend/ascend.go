// adapters/ascend provides a Chronos adapter for the Ascend weightlifting coaching platform.
package ascend

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/felixgeelhaar/chronos/internal/adapter"
	"github.com/felixgeelhaar/chronos/pkg/vector"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// Source implements adapter.Source for Ascend PostgreSQL databases.
type Source struct {
	db *sql.DB
}

// NewSource creates an Ascend adapter from a PostgreSQL connection string.
func NewSource(connStr string) (*Source, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("open ascend db: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping ascend db: %w", err)
	}
	return &Source{db: db}, nil
}

// Name returns the adapter identifier.
func (s *Source) Name() string {
	return "ascend"
}

// Fetch retrieves athlete training states from Ascend.
// cfg must contain "coach_id" (UUID of the coach whose athletes to fetch).
func (s *Source) Fetch(ctx context.Context, cfg map[string]string) ([]vector.EntityState, error) {
	coachIDStr, ok := cfg["coach_id"]
	if !ok {
		return nil, fmt.Errorf("ascend adapter requires coach_id in config")
	}

	coachID, err := uuid.Parse(coachIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid coach_id: %w", err)
	}

	// Query: get latest week per athlete with computed metrics
	// This is a read-only query against Ascend's database
	rows, err := s.db.QueryContext(ctx, `
		SELECT 
			a.id as athlete_id,
			apw.year,
			apw.week,
			apw.phase_type,
			apw.total_tonnes,
			apw.k1_tonnes,
			apw.k2_tonnes,
			apw.k3_tonnes,
			apw.k4_tonnes,
			apw.k5_tonnes,
			apw.k6_tonnes,
			apw.k7_tonnes,
			a.bodyweight_kg,
			apw.maz_focus
		FROM athletes a
		JOIN annual_plan_weeks apw ON a.id = apw.athlete_id
		WHERE a.coach_id = $1
		ORDER BY a.id, apw.year, apw.week
	`, coachID)
	if err != nil {
		return nil, fmt.Errorf("query ascend: %w", err)
	}
	defer rows.Close()

	var states []vector.EntityState
	for rows.Next() {
		var r struct {
			AthleteID    uuid.UUID
			Year         int
			Week         int
			PhaseType    string
			TotalTonnes  float64
			K1Tonnes     float64
			K2Tonnes     float64
			K3Tonnes     float64
			K4Tonnes     float64
			K5Tonnes     float64
			K6Tonnes     float64
			K7Tonnes     float64
			BodyweightKg float64
			MazFocus     sql.NullString
		}

		if err := rows.Scan(
			&r.AthleteID, &r.Year, &r.Week, &r.PhaseType,
			&r.TotalTonnes, &r.K1Tonnes, &r.K2Tonnes, &r.K3Tonnes,
			&r.K4Tonnes, &r.K5Tonnes, &r.K6Tonnes, &r.K7Tonnes,
			&r.BodyweightKg, &r.MazFocus,
		); err != nil {
			return nil, err
		}

		// Normalise tonnage by bodyweight
		normBW := r.BodyweightKg
		if normBW == 0 {
			normBW = 1 // avoid division by zero
		}

		features := []float64{
			r.TotalTonnes / normBW,
			r.K1Tonnes / normBW,
			r.K2Tonnes / normBW,
			r.K3Tonnes / normBW,
			r.K4Tonnes / normBW,
			r.K5Tonnes / normBW,
			r.K6Tonnes / normBW,
			r.K7Tonnes / normBW,
		}

		states = append(states, vector.EntityState{
			ID:        uuid.New(),
			EntityID:  r.AthleteID,
			ScopeID:   coachID,
			Timestamp: weekToTime(r.Year, r.Week),
			Features:  features,
			Labels: []string{
				"total_tonnes_per_kg", "k1_per_kg", "k2_per_kg", "k3_per_kg",
				"k4_per_kg", "k5_per_kg", "k6_per_kg", "k7_per_kg",
			},
			Meta: map[string]string{
				"phase_type": r.PhaseType,
				"maz_focus":  r.MazFocus.String,
				"year":       fmt.Sprintf("%d", r.Year),
				"week":       fmt.Sprintf("%d", r.Week),
			},
		})
	}

	return states, rows.Err()
}

// Close closes the database connection.
func (s *Source) Close() error {
	return s.db.Close()
}

func weekToTime(year, week int) time.Time {
	// Approximate: start of ISO week
	return time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC).
		Add(time.Duration(week-1) * 7 * 24 * time.Hour)
}

func init() {
	// Register the adapter so it can be loaded by name
	adapter.Register(&Source{})
}
