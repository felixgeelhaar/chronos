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

// Store implements store.Store for SQLite.
type Store struct {
	db *sql.DB
	q  *Queries
}

// NewStore creates a new SQLite store at the given path.
func NewStore(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite3", dbPath+"?_fk=on&_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	db.SetMaxOpenConns(1)
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

	for _, st := range states {
		features, _ := json.Marshal(st.Features)
		labels, _ := json.Marshal(st.Labels)
		meta, _ := json.Marshal(st.Meta)

		_, err := tx.ExecContext(ctx, `INSERT INTO entity_states (id, entity_id, scope_id, timestamp, features, labels, meta, adapter, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			st.ID.String(), st.EntityID.String(), st.ScopeID.String(),
			st.Timestamp.Format(time.RFC3339), string(features), string(labels), string(meta),
			adapter, time.Now().Format(time.RFC3339))
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// LoadEntityStates retrieves states for a scope.
func (s *Store) LoadEntityStates(ctx context.Context, scopeID uuid.UUID) ([]vector.EntityState, error) {
	rows, err := s.q.GetEntityStatesByScope(ctx, scopeID.String())
	if err != nil {
		return nil, err
	}

	states := make([]vector.EntityState, len(rows))
	for i, row := range rows {
		st, err := parseEntityState(row)
		if err != nil {
			return nil, err
		}
		states[i] = st
	}
	return states, nil
}

// LoadEntityStatesByEntity retrieves states for a specific entity.
func (s *Store) LoadEntityStatesByEntity(ctx context.Context, entityID uuid.UUID) ([]vector.EntityState, error) {
	rows, err := s.q.GetEntityStatesByEntity(ctx, entityID.String())
	if err != nil {
		return nil, err
	}

	states := make([]vector.EntityState, len(rows))
	for i, row := range rows {
		st, err := parseEntityState(row)
		if err != nil {
			return nil, err
		}
		states[i] = st
	}
	return states, nil
}

// DeleteOldEntityStates removes states older than the given threshold.
func (s *Store) DeleteOldEntityStates(ctx context.Context, before string, adapter string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM entity_states WHERE timestamp < ? AND adapter = ?", before, adapter)
	return err
}

// SaveInsight persists an insight.
func (s *Store) SaveInsight(ctx context.Context, in insight.Insight) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var validUntil string
	if in.ValidUntil != nil {
		validUntil = in.ValidUntil.Format(time.RFC3339)
	}

	_, err = tx.ExecContext(ctx, `INSERT INTO insights (id, scope_id, type, subject_entity, subject_time, sample_size, confidence, title, summary, suggestion, generated_at, valid_until)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		in.ID.String(), in.ScopeID.String(), in.Type, in.SubjectEntity.String(),
		in.SubjectTime.Format(time.RFC3339), in.SampleSize, in.Confidence,
		in.Title, in.Summary, in.Suggestion, in.GeneratedAt.Format(time.RFC3339), validUntil)
	if err != nil {
		return err
	}

	for _, c := range in.SimilarCases {
		_, err = tx.ExecContext(ctx, `INSERT INTO insight_cases (insight_id, entity_id, case_time, similarity) VALUES (?, ?, ?, ?)`,
			in.ID.String(), c.EntityID.String(), c.Time.Format(time.RFC3339), c.Similarity)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// LoadInsights retrieves active insights for a scope.
func (s *Store) LoadInsights(ctx context.Context, scopeID uuid.UUID) ([]insight.Insight, error) {
	rows, err := s.q.GetInsightsByScope(ctx, scopeID.String())
	if err != nil {
		return nil, err
	}

	insights := make([]insight.Insight, len(rows))
	for i, row := range rows {
		ins, err := parseInsight(row)
		if err != nil {
			return nil, err
		}
		insights[i] = ins
	}
	return insights, nil
}

// LoadInsightByID retrieves a single insight.
func (s *Store) LoadInsightByID(ctx context.Context, id uuid.UUID) (insight.Insight, error) {
	row, err := s.q.GetInsightByID(ctx, id.String())
	if err != nil {
		return insight.Insight{}, err
	}
	return parseInsight(row)
}

// DismissInsight marks an insight as dismissed.
func (s *Store) DismissInsight(ctx context.Context, insightID, dismissedBy uuid.UUID) error {
	_, err := s.db.ExecContext(ctx, "UPDATE insights SET dismissed_at = ?, dismissed_by = ? WHERE id = ?",
		time.Now().Format(time.RFC3339), dismissedBy.String(), insightID.String())
	return err
}

// SaveFeedback records feedback.
func (s *Store) SaveFeedback(ctx context.Context, insightID uuid.UUID, fb insight.Feedback) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO insight_feedback (insight_id, useful, applied, reason, created_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(insight_id) DO UPDATE SET useful = ?, applied = ?, reason = ?, created_at = ?`,
		insightID.String(), boolToInt(fb.Useful), boolToInt(fb.Applied), fb.Reason, fb.At.Format(time.RFC3339),
		boolToInt(fb.Useful), boolToInt(fb.Applied), fb.Reason, fb.At.Format(time.RFC3339))
	return err
}

// LoadFeedback retrieves feedback for an insight.
func (s *Store) LoadFeedback(ctx context.Context, insightID uuid.UUID) (insight.Feedback, error) {
	row, err := s.q.GetFeedbackByInsight(ctx, insightID.String())
	if err == sql.ErrNoRows {
		return insight.Feedback{}, nil
	}
	if err != nil {
		return insight.Feedback{}, err
	}

	return insight.Feedback{
		Useful:  row.Useful.Int64 == 1 && row.Useful.Valid,
		Applied: row.Applied.Int64 == 1 && row.Applied.Valid,
		Reason:  row.Reason,
		At:      time.Time{},
	}, nil
}

// CountEntityStates returns the count of entity states for an adapter.
func (s *Store) CountEntityStates(ctx context.Context, adapter string) (int64, error) {
	return s.q.CountEntityStates(ctx, adapter)
}

// CountInsights returns the count of insights for a scope.
func (s *Store) CountInsights(ctx context.Context, scopeID uuid.UUID) (int64, error) {
	return s.q.CountInsights(ctx, scopeID.String())
}

// --- helpers ---

func parseEntityState(row EntityState) (vector.EntityState, error) {
	st := vector.EntityState{}

	st.ID, _ = uuid.Parse(row.ID)
	st.EntityID, _ = uuid.Parse(row.EntityID)
	st.ScopeID, _ = uuid.Parse(row.ScopeID)
	st.Timestamp, _ = time.Parse(time.RFC3339, row.Timestamp)
	_ = json.Unmarshal([]byte(row.Features), &st.Features)
	_ = json.Unmarshal([]byte(row.Labels), &st.Labels)

	return st, nil
}

func parseInsight(row Insight) (insight.Insight, error) {
	in := insight.Insight{
		SampleSize:  int(row.SampleSize),
		Confidence: row.Confidence,
		Title:      row.Title,
		Summary:    row.Summary,
		Suggestion:  row.Suggestion,
	}
	in.ID, _ = uuid.Parse(row.ID)
	in.ScopeID, _ = uuid.Parse(row.ScopeID)
	in.Type = row.Type
	in.SubjectEntity, _ = uuid.Parse(row.SubjectEntity)
	in.SubjectTime, _ = time.Parse(time.RFC3339, row.SubjectTime)
	in.GeneratedAt, _ = time.Parse(time.RFC3339, row.GeneratedAt)

	if row.ValidUntil != "" {
		t, _ := time.Parse(time.RFC3339, row.ValidUntil)
		in.ValidUntil = &t
	}
	if row.DismissedAt != "" {
		t, _ := time.Parse(time.RFC3339, row.DismissedAt)
		in.DismissedAt = &t
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
	schema := `
CREATE TABLE IF NOT EXISTS entity_states (
    id TEXT PRIMARY KEY,
    entity_id TEXT NOT NULL,
    scope_id TEXT NOT NULL,
    timestamp TEXT NOT NULL,
    features TEXT NOT NULL,
    labels TEXT DEFAULT '[]',
    meta TEXT DEFAULT '{}',
    adapter TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_entity_states_scope ON entity_states(scope_id);
CREATE INDEX IF NOT EXISTS idx_entity_states_entity ON entity_states(entity_id);

CREATE TABLE IF NOT EXISTS insights (
    id TEXT PRIMARY KEY,
    scope_id TEXT NOT NULL,
    type TEXT NOT NULL,
    subject_entity TEXT NOT NULL,
    subject_time TEXT NOT NULL,
    sample_size INTEGER NOT NULL,
    confidence REAL NOT NULL,
    title TEXT NOT NULL,
    summary TEXT NOT NULL,
    suggestion TEXT,
    generated_at TEXT NOT NULL,
    valid_until TEXT,
    dismissed_at TEXT,
    dismissed_by TEXT
);
CREATE INDEX IF NOT EXISTS idx_insights_scope ON insights(scope_id);

CREATE TABLE IF NOT EXISTS insight_cases (
    insight_id TEXT NOT NULL,
    entity_id TEXT NOT NULL,
    case_time TEXT NOT NULL,
    similarity REAL NOT NULL,
    outcome_diff REAL,
    FOREIGN KEY (insight_id) REFERENCES insights(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS insight_feedback (
    insight_id TEXT PRIMARY KEY,
    useful INTEGER,
    applied INTEGER,
    reason TEXT,
    created_at TEXT NOT NULL,
    FOREIGN KEY (insight_id) REFERENCES insights(id) ON DELETE CASCADE
);
`
	_, err := db.ExecContext(context.Background(), schema)
	return err
}