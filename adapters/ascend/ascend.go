// Package ascend provides a Chronos adapter for the Ascend weightlifting
// coaching platform. It maps athlete training-week records into the generic
// chronos.EntityState shape so the engine can detect patterns across
// athletes within a coach's roster.
//
// Feature ordering: per-bodyweight tonnages for total and seven Klein zones
// (k1..k7), followed by total tonnage as the outcome metric (last element,
// per the engine convention).
package ascend

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/felixgeelhaar/chronos"
	"github.com/google/uuid"

	// pgx-stdlib registers the "pgx" sql.DB driver and replaces lib/pq
	// so the entire codebase shares one Postgres driver.
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Source is a chronos.Source backed by an Ascend PostgreSQL database.
type Source struct {
	db *sql.DB
}

// NewSource creates an Ascend adapter from a PostgreSQL connection string.
func NewSource(connStr string) (*Source, error) {
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, fmt.Errorf("ascend: open db: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ascend: ping db: %w", err)
	}
	return &Source{db: db}, nil
}

// Name returns the stable adapter identifier.
func (s *Source) Name() string { return "ascend" }

// Close releases the underlying database connection.
func (s *Source) Close() error {
	if s.db == nil {
		return nil
	}
	return s.db.Close()
}

// Fetch retrieves athlete training states for a coach. cfg["coach_id"] is
// required and must parse as a UUID.
func (s *Source) Fetch(ctx context.Context, cfg map[string]string) ([]chronos.EntityState, error) {
	if s.db == nil {
		return nil, fmt.Errorf("ascend: source not initialised; call NewSource")
	}
	coachIDStr, ok := cfg["coach_id"]
	if !ok {
		return nil, fmt.Errorf("ascend: coach_id required")
	}
	coachID, err := uuid.Parse(coachIDStr)
	if err != nil {
		return nil, fmt.Errorf("ascend: invalid coach_id: %w", err)
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT
			a.id, apw.year, apw.week, apw.phase_type,
			apw.total_tonnes, apw.k1_tonnes, apw.k2_tonnes, apw.k3_tonnes,
			apw.k4_tonnes, apw.k5_tonnes, apw.k6_tonnes, apw.k7_tonnes,
			a.bodyweight_kg, apw.maz_focus
		FROM athletes a
		JOIN annual_plan_weeks apw ON a.id = apw.athlete_id
		WHERE a.coach_id = $1
		ORDER BY a.id, apw.year, apw.week
	`, coachID)
	if err != nil {
		return nil, fmt.Errorf("ascend: query: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var states []chronos.EntityState
	for rows.Next() {
		var r row
		if err := rows.Scan(
			&r.athleteID, &r.year, &r.week, &r.phaseType,
			&r.totalTonnes, &r.k1, &r.k2, &r.k3, &r.k4, &r.k5, &r.k6, &r.k7,
			&r.bodyweight, &r.mazFocus,
		); err != nil {
			return nil, fmt.Errorf("ascend: scan: %w", err)
		}
		states = append(states, r.toEntityState(coachID))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ascend: rows: %w", err)
	}
	return states, nil
}

type row struct {
	athleteID                       uuid.UUID
	year, week                      int
	phaseType                       string
	totalTonnes, k1, k2, k3, k4, k5 float64
	k6, k7                          float64
	bodyweight                      float64
	mazFocus                        sql.NullString
}

func (r row) toEntityState(scope uuid.UUID) chronos.EntityState {
	bw := r.bodyweight
	if bw == 0 {
		bw = 1
	}
	// Last feature is the outcome metric per the engine convention; we use
	// total tonnage per bodyweight, since "more total work absorbed" is the
	// directional outcome a coach cares about for an aggregated week.
	features := []float64{
		r.k1 / bw, r.k2 / bw, r.k3 / bw, r.k4 / bw,
		r.k5 / bw, r.k6 / bw, r.k7 / bw,
		r.totalTonnes / bw,
	}
	labels := []string{"k1_per_kg", "k2_per_kg", "k3_per_kg", "k4_per_kg", "k5_per_kg", "k6_per_kg", "k7_per_kg", "total_per_kg"}
	return chronos.EntityState{
		ID:        uuid.New(),
		EntityID:  r.athleteID,
		ScopeID:   scope,
		Timestamp: weekStart(r.year, r.week),
		Features:  features,
		Labels:    labels,
		Meta: map[string]string{
			"phase_type": r.phaseType,
			"maz_focus":  r.mazFocus.String,
			"year":       fmt.Sprintf("%d", r.year),
			"week":       fmt.Sprintf("%d", r.week),
		},
	}
}

// weekStart returns the start of the given ISO year/week pair. The
// approximation (year-start + (week-1)*7d) is acceptable here because the
// engine compares timestamps but does not reason about calendar boundaries.
func weekStart(year, week int) time.Time {
	return time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC).
		Add(time.Duration(week-1) * 7 * 24 * time.Hour)
}

// Register the adapter so binaries that import this package make it
// available via chronos.Get("ascend"). Construction of the underlying *sql.DB
// is deferred to the caller via NewSource — the registered instance has a
// nil db and serves only to advertise the name; CLIs should call NewSource
// and re-Register if they need the operational source.
//
// In practice the chronos CLI today wires the ascend source explicitly. This
// init() makes a no-arg "name available" registration so chronos.Adapters()
// reports it; calling Fetch on the registered instance returns an error.
func init() { chronos.Register(&Source{}) }
