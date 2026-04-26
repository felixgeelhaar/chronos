// Package sqlite provides a SQLite-backed implementation of store.Store.
package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/felixgeelhaar/chronos/pkg/insight"
	"github.com/felixgeelhaar/chronos/pkg/vector"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

//go:generate sqlc generate

// Store implements store.Store for SQLite.
type Store struct {
	db *sql.DB
	q  *Queries
}

// New creates a new SQLite store at the given path.
func New(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite3", dbPath+"?_fk=on&_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	db.SetMaxOpenConns(1) // SQLite works best with single writer
	db.SetConnMaxLifetime(0)

	if err := migrate(db); err != nil {
		return nil, fmt.Errorf("migrate sqlite: %w", err)
	}

	return &Store{db: db, q: New(db)}, nil
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

	qtx := s.q.WithTx(tx)
	for _, st := range states {
		features, _ := json.Marshal(st.Features)
		labels, _ := json.Marshal(st.Labels)
		meta, _ := json.Marshal(st.Meta)

		if err := qtx.InsertEntityState(ctx, InsertEntityStateParams{
			ID:        st.ID,
			EntityID:  st.EntityID,
			ScopeID:   st.ScopeID,
			Timestamp: st.Timestamp.Format(time.RFC3339),
			Features:  string(features),
			Labels:    sql.NullString{String: string(labels), Valid: len(st.Labels) > 0},
			Meta:      sql.NullString{String: string(meta), Valid: len(st.Meta) > 0},
			Adapter:   adapter,
			CreatedAt: time.Now().Format(time.RFC3339),
		}); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// LoadEntityStates retrieves states for a scope.
func (s *Store) LoadEntityStates(ctx context.Context, scopeID uuid.UUID) ([]vector.EntityState, error) {
	rows, err := s.q.GetEntityStatesByScope(ctx, scopeID)
	if err != nil {
		return nil, err
	}

	var states []vector.EntityState
	for _, row := range rows {
		st, err := parseEntityState(row)
		if err != nil {
			return nil, err
		}
		states = append(states, st)
	}
	return states, nil
}

// LoadEntityStatesByEntity retrieves states for a specific entity.
func (s *Store) LoadEntityStatesByEntity(ctx context.Context, entityID uuid.UUID) ([]vector.EntityState, error) {
	rows, err := s.q.GetEntityStatesByEntity(ctx, entityID)
	if err != nil {
		return nil, err
	}

	var states []vector.EntityState
	for _, row := range rows {
		st, err := parseEntityStateRow(row)
		if err != nil {
			return nil, err
		}
		states = append(states, st)
	}
	return states, nil
}

// DeleteOldEntityStates removes states older than the given threshold.
func (s *Store) DeleteOldEntityStates(ctx context.Context, before string, adapter string) error {
	return s.q.DeleteOldEntityStates(ctx, DeleteOldEntityStatesParams{
		Timestamp: before,
		Adapter:   adapter,
	})
}

// SaveInsight persists an insight.
func (s *Store) SaveInsight(ctx context.Context, in insight.Insight) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	qtx := s.q.WithTx(tx)

	var validUntil sql.NullString
	if in.ValidUntil != nil {
		validUntil = sql.NullString{String: in.ValidUntil.Format(time.RFC3339), Valid: true}
	}

	if err := qtx.InsertInsight(ctx, InsertInsightParams{
		ID:            in.ID,
		ScopeID:       in.ScopeID,
		Type:          in.Type,
		SubjectEntity: in.SubjectEntity,
		SubjectTime:   in.SubjectTime.Format(time.RFC3339),
		SampleSize:    int64(in.SampleSize),
		Confidence:    in.Confidence,
		Title:         in.Title,
		Summary:       in.Summary,
		Suggestion:    sql.NullString{String: in.Suggestion, Valid: in.Suggestion != ""},
		GeneratedAt:   in.GeneratedAt.Format(time.RFC3339),
		ValidUntil:    validUntil,
	}); err != nil {
		return err
	}

	for _, c := range in.SimilarCases {
		if err := qtx.InsertInsightCase(ctx, InsertInsightCaseParams{
			InsightID:   in.ID,
			EntityID:    c.EntityID,
			CaseTime:    c.Time.Format(time.RFC3339),
			Similarity:  c.Similarity,
			OutcomeDiff: sql.NullFloat64{Float64: c.OutcomeDiff, Valid: true},
		}); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// LoadInsights retrieves active insights for a scope.
func (s *Store) LoadInsights(ctx context.Context, scopeID uuid.UUID) ([]insight.Insight, error) {
	rows, err := s.q.GetInsightsByScope(ctx, scopeID)
	if err != nil {
		return nil, err
	}

	var insights []insight.Insight
	for _, row := range rows {
		ins, err := parseInsight(row)
		if err != nil {
			return nil, err
		}
		insights = append(insights, ins)
	}
	return insights, nil
}

// LoadInsightByID retrieves a single insight.
func (s *Store) LoadInsightByID(ctx context.Context, id uuid.UUID) (insight.Insight, error) {
	row, err := s.q.GetInsightByID(ctx, id)
	if err != nil {
		return insight.Insight{}, err
	}
	return parseInsight(row)
}

// DismissInsight marks an insight as dismissed.
func (s *Store) DismissInsight(ctx context.Context, insightID, dismissedBy uuid.UUID) error {
	return s.q.DismissInsight(ctx, DismissInsightParams{
		DismissedAt: sql.NullString{String: time.Now().Format(time.RFC3339), Valid: true},
		DismissedBy: sql.NullString{String: dismissedBy.String(), Valid: dismissedBy != uuid.Nil},
		ID:          insightID,
	})
}

// SaveFeedback records feedback.
func (s *Store) SaveFeedback(ctx context.Context, insightID uuid.UUID, fb insight.Feedback) error {
	if fb.At.IsZero() {
		return fmt.Errorf("feedback timestamp required")
	}

	return s.q.UpsertFeedback(ctx, UpsertFeedbackParams{
		InsightID: insightID,
		Useful:    sql.NullInt64{Int64: boolToInt(fb.Useful), Valid: true},
		Applied:   sql.NullInt64{Int64: boolToInt(fb.Applied), Valid: true},
		Reason:    sql.NullString{String: fb.Reason, Valid: fb.Reason != ""},
		CreatedAt: fb.At.Format(time.RFC3339),
	})
}

// LoadFeedback retrieves feedback for an insight.
func (s *Store) LoadFeedback(ctx context.Context, insightID uuid.UUID) (insight.Feedback, error) {
	row, err := s.q.GetFeedbackByInsight(ctx, insightID)
	if err != nil {
		if err == sql.ErrNoRows {
			return insight.Feedback{}, nil
		}
		return insight.Feedback{}, err
	}

	fb := insight.Feedback{
		Useful:  row.Useful.Valid && row.Useful.Int64 == 1,
		Applied: row.Applied.Valid && row.Applied.Int64 == 1,
		Reason:  row.Reason.String,
	}
	if t, err := time.Parse(time.RFC3339, row.CreatedAt); err == nil {
		fb.At = t
	}
	return fb, nil
}

// CountEntityStates returns the count of entity states for an adapter.
func (s *Store) CountEntityStates(ctx context.Context, adapter string) (int64, error) {
	return s.q.CountEntityStates(ctx, adapter)
}

// CountInsights returns the count of insights for a scope.
func (s *Store) CountInsights(ctx context.Context, scopeID uuid.UUID) (int64, error) {
	return s.q.CountInsights(ctx, scopeID)
}

// --- helpers ---

func parseEntityState(row GetEntityStatesByScopeRow) (vector.EntityState, error) {
	return parseEntityStateRow(row)
}

func parseEntityStateRow(row interface{}) (vector.EntityState, error) {
	var st vector.EntityState

	switch r := row.(type) {
	case GetEntityStatesByScopeRow:
		st.ID = r.ID
		st.EntityID = r.EntityID
		st.ScopeID = r.ScopeID
		t, _ := time.Parse(time.RFC3339, r.Timestamp)
		st.Timestamp = t
		_ = json.Unmarshal([]byte(r.Features), &st.Features)
		if r.Labels.Valid {
			_ = json.Unmarshal([]byte(r.Labels.String), &st.Labels)
		}
		if r.Meta.Valid {
			_ = json.Unmarshal([]byte(r.Meta.String), &st.Meta)
		}
	case GetEntityStatesByEntityRow:
		st.ID = r.ID
		st.EntityID = r.EntityID
		st.ScopeID = r.ScopeID
		t, _ := time.Parse(time.RFC3339, r.Timestamp)
		st.Timestamp = t
		_ = json.Unmarshal([]byte(r.Features), &st.Features)
		if r.Labels.Valid {
			_ = json.Unmarshal([]byte(r.Labels.String), &st.Labels)
		}
		if r.Meta.Valid {
			_ = json.Unmarshal([]byte(r.Meta.String), &st.Meta)
		}
	}

	return st, nil
}

func parseInsight(row interface{}) (insight.Insight, error) {
	var in insight.Insight

	switch r := row.(type) {
	case GetInsightsByScopeRow:
		in.ID = r.ID
		in.ScopeID = r.ScopeID
		in.Type = r.Type
		in.SubjectEntity = r.SubjectEntity
		t, _ := time.Parse(time.RFC3339, r.SubjectTime)
		in.SubjectTime = t
		in.SampleSize = int(r.SampleSize)
		in.Confidence = r.Confidence
		in.Title = r.Title
		in.Summary = r.Summary
		if r.Suggestion.Valid {
			in.Suggestion = r.Suggestion.String
		}
		gt, _ := time.Parse(time.RFC3339, r.GeneratedAt)
		in.GeneratedAt = gt
	case GetInsightByIDRow:
		in.ID = r.ID
		in.ScopeID = r.ScopeID
		in.Type = r.Type
		in.SubjectEntity = r.SubjectEntity
		t, _ := time.Parse(time.RFC3339, r.SubjectTime)
		in.SubjectTime = t
		in.SampleSize = int(r.SampleSize)
		in.Confidence = r.Confidence
		in.Title = r.Title
		in.Summary = r.Summary
		if r.Suggestion.Valid {
			in.Suggestion = r.Suggestion.String
		}
		gt, _ := time.Parse(time.RFC3339, r.GeneratedAt)
		in.GeneratedAt = gt
	}

	return in, nil
}

func boolToInt(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

func migrate(db *sql.DB) error {
	schema := `CREATE TABLE IF NOT EXISTS entity_states (
    id TEXT PRIMARY KEY,
    entity_id TEXT NOT NULL,
    scope_id TEXT NOT NULL,
    timestamp TEXT NOT NULL,
    features TEXT NOT NULL,
    labels TEXT,
    meta TEXT,
    adapter TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_entity_states_scope ON entity_states(scope_id);
CREATE INDEX IF NOT EXISTS idx_entity_states_entity ON entity_states(entity_id);
CREATE INDEX IF NOT EXISTS idx_entity_states_time ON entity_states(timestamp);

CREATE TABLE IF NOT EXISTS similarities (
    id TEXT PRIMARY KEY,
    state_a_id TEXT NOT NULL,
    state_b_id TEXT NOT NULL,
    similarity REAL NOT NULL CHECK (similarity >= -1 AND similarity <= 1),
    computed_at TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE(state_a_id, state_b_id)
);

CREATE TABLE IF NOT EXISTS insights (
    id TEXT PRIMARY KEY,
    scope_id TEXT NOT NULL,
    type TEXT NOT NULL,
    subject_entity TEXT NOT NULL,
    subject_time TEXT NOT NULL,
    sample_size INTEGER NOT NULL,
    confidence REAL NOT NULL CHECK (confidence >= 0 AND confidence <= 1),
    title TEXT NOT NULL,
    summary TEXT NOT NULL,
    suggestion TEXT,
    generated_at TEXT NOT NULL,
    valid_until TEXT,
    dismissed_at TEXT,
    dismissed_by TEXT
);
CREATE INDEX IF NOT EXISTS idx_insights_scope ON insights(scope_id);
CREATE INDEX IF NOT EXISTS idx_insights_active ON insights(scope_id, dismissed_at) WHERE dismissed_at IS NULL;

CREATE TABLE IF NOT EXISTS insight_cases (
    insight_id TEXT NOT NULL,
    entity_id TEXT NOT NULL,
    case_time TEXT NOT NULL,
    similarity REAL NOT NULL,
    outcome_diff REAL,
    FOREIGN KEY (insight_id) REFERENCES insights(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_insight_cases ON insight_cases(insight_id);

CREATE TABLE IF NOT EXISTS insight_feedback (
    insight_id TEXT PRIMARY KEY,
    useful INTEGER,
    applied INTEGER,
    reason TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (insight_id) REFERENCES insights(id) ON DELETE CASCADE
);
`
	_, err := db.Exec(schema)
	return err
}
