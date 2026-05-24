package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/felixgeelhaar/chronos/internal/domain"
	"github.com/felixgeelhaar/chronos/internal/ports"
	"github.com/google/uuid"
)

// SignalRepository persists signals and their evidence in PostgreSQL.
type SignalRepository struct{ conn *Conn }

// Save writes a signal and its evidence atomically.
func (r *SignalRepository) Save(ctx context.Context, sig domain.Signal) error {
	if err := sig.Validate(); err != nil {
		return err
	}
	tx, err := r.conn.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("signal save: begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	metricsJSON, _ := json.Marshal(sig.Metrics)
	_, err = tx.ExecContext(ctx, `
		INSERT INTO signals (id, scope_id, series_id, pattern, detected_at, window_start, window_end, strength, confidence, metrics)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			strength = EXCLUDED.strength,
			confidence = EXCLUDED.confidence,
			metrics = EXCLUDED.metrics
	`, sig.ID, sig.ScopeID, sig.Series, string(sig.Pattern),
		sig.DetectedAt, sig.Window.Start, sig.Window.End,
		sig.Strength, sig.Confidence, metricsJSON,
	)
	if err != nil {
		return fmt.Errorf("signal save: insert: %w", err)
	}
	for _, e := range sig.Evidence {
		evMetrics, _ := json.Marshal(e.Metrics)
		_, err := tx.ExecContext(ctx, `
			INSERT INTO signal_evidence (signal_id, series_id, time, kind, score, metrics)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, sig.ID, e.Series, e.Time, e.Kind, e.Score, evMetrics)
		if err != nil {
			return fmt.Errorf("signal save: insert evidence: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("signal save: commit: %w", err)
	}
	return nil
}

// List returns signals matching filter, ordered detected-at desc then
// confidence desc.
func (r *SignalRepository) List(ctx context.Context, filter ports.SignalFilter) ([]domain.Signal, error) {
	query, args := buildListQuery(filter)
	rows, err := r.conn.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("signal list: %w", err)
	}
	signals, err := scanSignals(rows)
	if err != nil {
		return nil, err
	}
	for i := range signals {
		ev, err := r.loadEvidence(ctx, signals[i].ID)
		if err != nil {
			return nil, err
		}
		signals[i].Evidence = ev
	}
	return signals, nil
}

// Get returns a single signal by ID, including its evidence.
func (r *SignalRepository) Get(ctx context.Context, id uuid.UUID) (domain.Signal, error) {
	row := r.conn.DB.QueryRowContext(ctx, `
		SELECT id, scope_id, series_id, pattern, detected_at, window_start, window_end, strength, confidence, metrics
		FROM signals WHERE id = $1
	`, id)
	sig, err := scanOneSignal(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Signal{}, domain.ErrSignalNotFound
	}
	if err != nil {
		return domain.Signal{}, fmt.Errorf("signal get: %w", err)
	}
	ev, err := r.loadEvidence(ctx, id)
	if err != nil {
		return domain.Signal{}, err
	}
	sig.Evidence = ev
	return sig, nil
}

// Count returns the number of signals matching filter.
func (r *SignalRepository) Count(ctx context.Context, filter ports.SignalFilter) (int64, error) {
	query, args := buildCountQuery(filter)
	var n int64
	if err := r.conn.DB.QueryRowContext(ctx, query, args...).Scan(&n); err != nil {
		return 0, fmt.Errorf("signal count: %w", err)
	}
	return n, nil
}

func (r *SignalRepository) loadEvidence(ctx context.Context, id uuid.UUID) ([]domain.Evidence, error) {
	rows, err := r.conn.DB.QueryContext(ctx, `
		SELECT series_id, time, kind, score, metrics
		FROM signal_evidence
		WHERE signal_id = $1
		ORDER BY score DESC
	`, id)
	if err != nil {
		return nil, fmt.Errorf("signal evidence: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var out []domain.Evidence
	for rows.Next() {
		var e domain.Evidence
		var metrics []byte
		if err := rows.Scan(&e.Series, &e.Time, &e.Kind, &e.Score, &metrics); err != nil {
			return nil, fmt.Errorf("scan evidence: %w", err)
		}
		if len(metrics) > 0 {
			_ = json.Unmarshal(metrics, &e.Metrics)
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func buildListQuery(f ports.SignalFilter) (string, []any) {
	const base = `SELECT id, scope_id, series_id, pattern, detected_at, window_start, window_end, strength, confidence, metrics FROM signals`
	where, args := buildWhere(f)
	q := base + where + " ORDER BY detected_at DESC, confidence DESC"
	if f.Limit > 0 {
		q += fmt.Sprintf(" LIMIT %d", f.Limit)
	}
	return q, args
}

func buildCountQuery(f ports.SignalFilter) (string, []any) {
	const base = `SELECT COUNT(*) FROM signals`
	where, args := buildWhere(f)
	return base + where, args
}

func buildWhere(f ports.SignalFilter) (string, []any) {
	var clauses []string
	var args []any
	add := func(predicate string, val any) {
		args = append(args, val)
		clauses = append(clauses, fmt.Sprintf("%s $%d", predicate, len(args)))
	}
	if f.ScopeID != uuid.Nil {
		add("scope_id =", f.ScopeID)
	}
	if len(f.ScopeIDs) > 0 {
		placeholders := make([]string, len(f.ScopeIDs))
		for i, id := range f.ScopeIDs {
			args = append(args, id)
			placeholders[i] = fmt.Sprintf("$%d", len(args))
		}
		clauses = append(clauses, "scope_id IN ("+strings.Join(placeholders, ",")+")")
	}
	if f.Series != nil {
		add("series_id =", *f.Series)
	}
	if f.Pattern != nil {
		add("pattern =", string(*f.Pattern))
	}
	if f.Since != nil {
		add("detected_at >=", *f.Since)
	}
	if f.Until != nil {
		add("detected_at <", *f.Until)
	}
	if f.MinConfidence != nil {
		add("confidence >=", *f.MinConfidence)
	}
	if len(clauses) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(clauses, " AND "), args
}

func scanSignals(rows *sql.Rows) ([]domain.Signal, error) {
	defer func() { _ = rows.Close() }()
	var out []domain.Signal
	for rows.Next() {
		sig, err := scanSignalRow(rows.Scan)
		if err != nil {
			return nil, err
		}
		out = append(out, sig)
	}
	return out, rows.Err()
}

func scanOneSignal(row *sql.Row) (domain.Signal, error) {
	return scanSignalRow(row.Scan)
}

func scanSignalRow(scan func(...any) error) (domain.Signal, error) {
	var (
		sig         domain.Signal
		patternStr  string
		metricsJSON []byte
	)
	if err := scan(
		&sig.ID, &sig.ScopeID, &sig.Series, &patternStr,
		&sig.DetectedAt, &sig.Window.Start, &sig.Window.End,
		&sig.Strength, &sig.Confidence, &metricsJSON,
	); err != nil {
		return domain.Signal{}, err
	}
	sig.Pattern = domain.PatternType(patternStr)
	if len(metricsJSON) > 0 {
		_ = json.Unmarshal(metricsJSON, &sig.Metrics)
	}
	return sig, nil
}
