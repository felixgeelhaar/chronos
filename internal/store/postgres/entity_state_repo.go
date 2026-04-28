package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/felixgeelhaar/chronos"
	"github.com/google/uuid"
)

// EntityStateRepository persists entity states in PostgreSQL.
type EntityStateRepository struct{ conn *Conn }

// Ingest persists a single observation.
func (r *EntityStateRepository) Ingest(ctx context.Context, adapterName string, state chronos.EntityState) error {
	if err := state.Validate(); err != nil {
		return fmt.Errorf("entity_state ingest: validate: %w", err)
	}
	return r.upsert(ctx, r.conn.DB, adapterName, state)
}

// Save upserts a batch of observations inside a single transaction.
func (r *EntityStateRepository) Save(ctx context.Context, adapterName string, states []chronos.EntityState) error {
	tx, err := r.conn.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("entity_state save: begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	for _, s := range states {
		if err := s.Validate(); err != nil {
			return fmt.Errorf("entity_state save: validate %s: %w", s.ID, err)
		}
		if err := r.upsert(ctx, tx, adapterName, s); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("entity_state save: commit: %w", err)
	}
	return nil
}

type execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func (r *EntityStateRepository) upsert(ctx context.Context, ex execer, adapterName string, s chronos.EntityState) error {
	featuresJSON, _ := json.Marshal(s.Features)
	labelsJSON, _ := json.Marshal(s.Labels)
	metaJSON, _ := json.Marshal(s.Meta)
	_, err := ex.ExecContext(ctx, `
		INSERT INTO entity_states (id, entity_id, scope_id, timestamp, features, labels, meta, adapter, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET
			features = EXCLUDED.features,
			labels = EXCLUDED.labels,
			meta = EXCLUDED.meta,
			adapter = EXCLUDED.adapter
	`,
		s.ID, s.EntityID, s.ScopeID,
		s.Timestamp,
		featuresJSON, labelsJSON, metaJSON,
		adapterName, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("entity_state upsert %s: %w", s.ID, err)
	}
	return nil
}

// ListByScope returns all entity states for a scope, most recent first.
func (r *EntityStateRepository) ListByScope(ctx context.Context, scopeID uuid.UUID) ([]chronos.EntityState, error) {
	rows, err := r.conn.DB.QueryContext(ctx, `
		SELECT id, entity_id, scope_id, timestamp, features, labels, meta
		FROM entity_states
		WHERE scope_id = $1
		ORDER BY timestamp DESC
	`, scopeID)
	if err != nil {
		return nil, fmt.Errorf("entity_state list by scope: %w", err)
	}
	return scanEntityStates(rows)
}

// ListByEntity returns all observations of an entity, most recent first.
func (r *EntityStateRepository) ListByEntity(ctx context.Context, entityID uuid.UUID) ([]chronos.EntityState, error) {
	rows, err := r.conn.DB.QueryContext(ctx, `
		SELECT id, entity_id, scope_id, timestamp, features, labels, meta
		FROM entity_states
		WHERE entity_id = $1
		ORDER BY timestamp DESC
	`, entityID)
	if err != nil {
		return nil, fmt.Errorf("entity_state list by entity: %w", err)
	}
	return scanEntityStates(rows)
}

// DeleteOlderThan removes states observed before cutoff for adapterName.
func (r *EntityStateRepository) DeleteOlderThan(ctx context.Context, cutoff time.Time, adapterName string) error {
	if _, err := r.conn.DB.ExecContext(ctx,
		`DELETE FROM entity_states WHERE timestamp < $1 AND adapter = $2`,
		cutoff, adapterName,
	); err != nil {
		return fmt.Errorf("entity_state delete old: %w", err)
	}
	return nil
}

// Count returns the number of states stored under adapterName.
func (r *EntityStateRepository) Count(ctx context.Context, adapterName string) (int64, error) {
	var n int64
	if err := r.conn.DB.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM entity_states WHERE adapter = $1`, adapterName,
	).Scan(&n); err != nil {
		return 0, fmt.Errorf("entity_state count: %w", err)
	}
	return n, nil
}

func scanEntityStates(rows *sql.Rows) ([]chronos.EntityState, error) {
	defer func() { _ = rows.Close() }()
	var out []chronos.EntityState
	for rows.Next() {
		var s chronos.EntityState
		var features, labels, meta []byte
		if err := rows.Scan(&s.ID, &s.EntityID, &s.ScopeID, &s.Timestamp, &features, &labels, &meta); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		_ = json.Unmarshal(features, &s.Features)
		if len(labels) > 0 {
			_ = json.Unmarshal(labels, &s.Labels)
		}
		if len(meta) > 0 {
			_ = json.Unmarshal(meta, &s.Meta)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}
