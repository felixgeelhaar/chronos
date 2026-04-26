-- name: InsertEntityState :exec
INSERT INTO entity_states (id, entity_id, scope_id, timestamp, features, labels, meta, adapter, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetEntityStatesByScope :many
SELECT * FROM entity_states
WHERE scope_id = ?
ORDER BY timestamp DESC;

-- name: GetEntityStatesByEntity :many
SELECT * FROM entity_states
WHERE entity_id = ?
ORDER BY timestamp DESC;

-- name: InsertSimilarity :exec
INSERT INTO similarities (id, state_a_id, state_b_id, similarity, computed_at)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT(state_a_id, state_b_id) DO UPDATE SET
    similarity = excluded.similarity,
    computed_at = excluded.computed_at;

-- name: GetSimilaritiesByState :many
SELECT * FROM similarities
WHERE state_a_id = ? OR state_b_id = ?
ORDER BY similarity DESC;

-- name: InsertInsight :exec
INSERT INTO insights (id, scope_id, type, subject_entity, subject_time, sample_size, confidence, title, summary, suggestion, generated_at, valid_until)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: InsertInsightCase :exec
INSERT INTO insight_cases (insight_id, entity_id, case_time, similarity, outcome_diff)
VALUES (?, ?, ?, ?, ?);

-- name: GetInsightsByScope :many
SELECT * FROM insights
WHERE scope_id = ? AND dismissed_at IS NULL
ORDER BY confidence DESC, generated_at DESC;

-- name: GetInsightByID :one
SELECT * FROM insights WHERE id = ?;

-- name: DismissInsight :exec
UPDATE insights SET dismissed_at = ?, dismissed_by = ? WHERE id = ?;

-- name: UpsertFeedback :exec
INSERT INTO insight_feedback (insight_id, useful, applied, reason, created_at)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT(insight_id) DO UPDATE SET
    useful = excluded.useful,
    applied = excluded.applied,
    reason = excluded.reason,
    created_at = excluded.created_at;

-- name: GetFeedbackByInsight :one
SELECT * FROM insight_feedback WHERE insight_id = ?;

-- name: DeleteOldEntityStates :exec
DELETE FROM entity_states
WHERE timestamp < ? AND adapter = ?;

-- name: CountEntityStates :one
SELECT COUNT(*) FROM entity_states WHERE adapter = ?;

-- name: CountInsights :one
SELECT COUNT(*) FROM insights WHERE scope_id = ?;
