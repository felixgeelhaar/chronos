// Package postgres provides a PostgreSQL-backed implementation of store.Store.
package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/felixgeelhaar/chronos/pkg/insight"
	"github.com/felixgeelhaar/chronos/pkg/vector"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// Store implements store.Store for PostgreSQL.
type Store struct {
	db *sql.DB
}

// New creates a new PostgreSQL store.
func New(connStr string) (*Store, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := migrate(db); err != nil {
		return nil, fmt.Errorf("migrate postgres: %w", err)
	}

	return &Store{db: db}, nil
}

// Close closes the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

// SaveEntityStates persists entity states.
func (s *Store) SaveEntityStates(ctx context.Context, adapter string, states []vector.EntityState) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO entity_states (id, entity_id, scope_id, timestamp, features, labels, meta, adapter, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET
			features = EXCLUDED.features,
			labels = EXCLUDED.labels,
			meta = EXCLUDED.meta,
			adapter = EXCLUDED.adapter
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, st := range states {
		features, _ := json.Marshal(st.Features)
		labels, _ := json.Marshal(st.Labels)
		meta, _ := json.Marshal(st.Meta)

		_, err := stmt.ExecContext(ctx,
			st.ID, st.EntityID, st.ScopeID,
			st.Timestamp.Format(time.RFC3339),
			features, labels, meta,
			adapter, time.Now().Format(time.RFC3339),
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// LoadEntityStates retrieves states for a scope.
func (s *Store) LoadEntityStates(ctx context.Context, scopeID uuid.UUID) ([]vector.EntityState, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, entity_id, scope_id, timestamp, features, labels, meta
		FROM entity_states
		WHERE scope_id = $1
		ORDER BY timestamp DESC
	`, scopeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanEntityStates(rows)
}

// LoadEntityStatesByEntity retrieves states for a specific entity.
func (s *Store) LoadEntityStatesByEntity(ctx context.Context, entityID uuid.UUID) ([]vector.EntityState, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, entity_id, scope_id, timestamp, features, labels, meta
		FROM entity_states
		WHERE entity_id = $1
		ORDER BY timestamp DESC
	`, entityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanEntityStates(rows)
}

// DeleteOldEntityStates removes states older than the given threshold.
func (s *Store) DeleteOldEntityStates(ctx context.Context, before string, adapter string) error {
	_, err := s.db.ExecContext(ctx, `
		DELETE FROM entity_states
		WHERE timestamp < $1 AND adapter = $2
	`, before, adapter)
	return err
}

// SaveInsight persists an insight with cases.
func (s *Store) SaveInsight(ctx context.Context, in insight.Insight) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var validUntil *time.Time
	if in.ValidUntil != nil {
		validUntil = in.ValidUntil
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO insights (id, scope_id, type, subject_entity, subject_time, sample_size, confidence, title, summary, suggestion, generated_at, valid_until)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (id) DO UPDATE SET
			confidence = EXCLUDED.confidence,
			summary = EXCLUDED.summary,
			suggestion = EXCLUDED.suggestion
	`, in.ID, in.ScopeID, in.Type, in.SubjectEntity, in.SubjectTime.Format(time.RFC3339),
		in.SampleSize, in.Confidence, in.Title, in.Summary, in.Suggestion,
		in.GeneratedAt.Format(time.RFC3339), validUntil)
	if err != nil {
		return err
	}

	for _, c := range in.SimilarCases {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO insight_cases (insight_id, entity_id, case_time, similarity, outcome_diff)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT DO NOTHING
		`, in.ID, c.EntityID, c.Time.Format(time.RFC3339), c.Similarity, c.OutcomeDiff)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// LoadInsights retrieves active insights for a scope.
func (s *Store) LoadInsights(ctx context.Context, scopeID uuid.UUID) ([]insight.Insight, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, scope_id, type, subject_entity, subject_time, sample_size, confidence, title, summary, suggestion, generated_at, valid_until
		FROM insights
		WHERE scope_id = $1 AND dismissed_at IS NULL
		ORDER BY confidence DESC, generated_at DESC
	`, scopeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanInsights(rows)
}

// LoadInsightByID retrieves a single insight.
func (s *Store) LoadInsightByID(ctx context.Context, id uuid.UUID) (insight.Insight, error) {
	var in insight.Insight
	var subjectTime, generatedAt time.Time
	var validUntil sql.NullTime
	var suggestion sql.NullString

	err := s.db.QueryRowContext(ctx, `
		SELECT id, scope_id, type, subject_entity, subject_time, sample_size, confidence, title, summary, suggestion, generated_at, valid_until
		FROM insights WHERE id = $1
	`, id).Scan(
		&in.ID, &in.ScopeID, &in.Type, &in.SubjectEntity, &subjectTime,
		&in.SampleSize, &in.Confidence, &in.Title, &in.Summary, &suggestion,
		&generatedAt, &validUntil,
	)
	if err != nil {
		return insight.Insight{}, err
	}

	in.SubjectTime = subjectTime
	in.GeneratedAt = generatedAt
	if validUntil.Valid {
		in.ValidUntil = &validUntil.Time
	}
	if suggestion.Valid {
		in.Suggestion = suggestion.String
	}

	return in, nil
}

// DismissInsight marks an insight as dismissed.
func (s *Store) DismissInsight(ctx context.Context, insightID, dismissedBy uuid.UUID) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE insights SET dismissed_at = $1, dismissed_by = $2 WHERE id = $3
	`, time.Now().Format(time.RFC3339), dismissedBy, insightID)
	return err
}

// SaveFeedback records feedback.
func (s *Store) SaveFeedback(ctx context.Context, insightID uuid.UUID, fb insight.Feedback) error {
	if fb.At.IsZero() {
		return fmt.Errorf("feedback timestamp required")
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO insight_feedback (insight_id, useful, applied, reason, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (insight_id) DO UPDATE SET
			useful = EXCLUDED.useful,
			applied = EXCLUDED.applied,
			reason = EXCLUDED.reason,
			created_at = EXCLUDED.created_at
	`, insightID, fb.Useful, fb.Applied, fb.Reason, fb.At.Format(time.RFC3339))
	return err
}

// LoadFeedback retrieves feedback.
func (s *Store) LoadFeedback(ctx context.Context, insightID uuid.UUID) (insight.Feedback, error) {
	var fb insight.Feedback
	var useful, applied sql.NullBool
	var reason sql.NullString
	var createdAt time.Time

	err := s.db.QueryRowContext(ctx, `
		SELECT useful, applied, reason, created_at FROM insight_feedback WHERE insight_id = $1
	`, insightID).Scan(&useful, &applied, &reason, &createdAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return insight.Feedback{}, nil
		}
		return insight.Feedback{}, err
	}

	if useful.Valid {
		fb.Useful = useful.Bool
	}
	if applied.Valid {
		fb.Applied = applied.Bool
	}
	fb.Reason = reason.String
	fb.At = createdAt
	return fb, nil
}

// CountEntityStates returns count for an adapter.
func (s *Store) CountEntityStates(ctx context.Context, adapter string) (int64, error) {
	var count int64
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM entity_states WHERE adapter = $1`, adapter).Scan(&count)
	return count, err
}

// CountInsights returns count for a scope.
func (s *Store) CountInsights(ctx context.Context, scopeID uuid.UUID) (int64, error) {
	var count int64
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM insights WHERE scope_id = $1`, scopeID).Scan(&count)
	return count, err
}

// --- helpers ---

func scanEntityStates(rows *sql.Rows) ([]vector.EntityState, error) {
	defer rows.Close()

	var states []vector.EntityState
	for rows.Next() {
		var st vector.EntityState
		var ts time.Time
		var labels, meta sql.NullString

		err := rows.Scan(&st.ID, &st.EntityID, &st.ScopeID, &ts, &st.Features, &labels, &meta)
		if err != nil {
			return nil, err
		}

		st.Timestamp = ts
		if labels.Valid {
			_ = json.Unmarshal([]byte(labels.String), &st.Labels)
		}
		if meta.Valid {
			_ = json.Unmarshal([]byte(meta.String), &st.Meta)
		}
		states = append(states, st)
	}
	return states, rows.Err()
}

func scanInsights(rows *sql.Rows) ([]insight.Insight, error) {
	defer rows.Close()

	var insights []insight.Insight
	for rows.Next() {
		var in insight.Insight
		var subjectTime, generatedAt time.Time
		var validUntil sql.NullTime
		var suggestion sql.NullString

		err := rows.Scan(
			&in.ID, &in.ScopeID, &in.Type, &in.SubjectEntity, &subjectTime,
			&in.SampleSize, &in.Confidence, &in.Title, &in.Summary, &suggestion,
			&generatedAt, &validUntil,
		)
		if err != nil {
			return nil, err
		}

		in.SubjectTime = subjectTime
		in.GeneratedAt = generatedAt
		if validUntil.Valid {
			in.ValidUntil = &validUntil.Time
		}
		if suggestion.Valid {
			in.Suggestion = suggestion.String
		}
		insights = append(insights, in)
	}
	return insights, rows.Err()
}

func migrate(db *sql.DB) error {
	schema := `CREATE TABLE IF NOT EXISTS entity_states (
    id UUID PRIMARY KEY,
    entity_id UUID NOT NULL,
    scope_id UUID NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    features JSONB NOT NULL,
    labels JSONB,
    meta JSONB,
    adapter TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_entity_states_scope ON entity_states(scope_id);
CREATE INDEX IF NOT EXISTS idx_entity_states_entity ON entity_states(entity_id);
CREATE INDEX IF NOT EXISTS idx_entity_states_time ON entity_states(timestamp);

CREATE TABLE IF NOT EXISTS similarities (
    id UUID PRIMARY KEY,
    state_a_id UUID NOT NULL,
    state_b_id UUID NOT NULL,
    similarity REAL NOT NULL CHECK (similarity >= -1 AND similarity <= 1),
    computed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(state_a_id, state_b_id)
);

CREATE TABLE IF NOT EXISTS insights (
    id UUID PRIMARY KEY,
    scope_id UUID NOT NULL,
    type TEXT NOT NULL,
    subject_entity UUID NOT NULL,
    subject_time TIMESTAMPTZ NOT NULL,
    sample_size INTEGER NOT NULL,
    confidence REAL NOT NULL CHECK (confidence >= 0 AND confidence <= 1),
    title TEXT NOT NULL,
    summary TEXT NOT NULL,
    suggestion TEXT,
    generated_at TIMESTAMPTZ NOT NULL,
    valid_until TIMESTAMPTZ,
    dismissed_at TIMESTAMPTZ,
    dismissed_by UUID
);
CREATE INDEX IF NOT EXISTS idx_insights_scope ON insights(scope_id);
CREATE INDEX IF NOT EXISTS idx_insights_active ON insights(scope_id, dismissed_at) WHERE dismissed_at IS NULL;

CREATE TABLE IF NOT EXISTS insight_cases (
    insight_id UUID NOT NULL REFERENCES insights(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL,
    case_time TIMESTAMPTZ NOT NULL,
    similarity REAL NOT NULL,
    outcome_diff REAL,
    PRIMARY KEY (insight_id, entity_id, case_time)
);
CREATE INDEX IF NOT EXISTS idx_insight_cases ON insight_cases(insight_id);

CREATE TABLE IF NOT EXISTS insight_feedback (
    insight_id UUID PRIMARY KEY REFERENCES insights(id) ON DELETE CASCADE,
    useful BOOLEAN,
    applied BOOLEAN,
    reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`
	_, err := db.Exec(schema)
	return err
}
