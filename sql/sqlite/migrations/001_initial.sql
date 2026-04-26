-- Schema migration 001: Initial schema for Chronos generic pattern detection

CREATE TABLE entity_states (
    id TEXT PRIMARY KEY,
    entity_id TEXT NOT NULL,
    scope_id TEXT NOT NULL,
    timestamp TEXT NOT NULL, -- RFC3339
    features TEXT NOT NULL,  -- JSON array of float64
    labels TEXT,             -- JSON array of feature names
    meta TEXT,               -- JSON object of adapter metadata
    adapter TEXT NOT NULL,   -- Source adapter name
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_entity_states_scope ON entity_states(scope_id);
CREATE INDEX idx_entity_states_entity ON entity_states(entity_id);
CREATE INDEX idx_entity_states_time ON entity_states(timestamp);
CREATE INDEX idx_entity_states_adapter ON entity_states(adapter);

CREATE TABLE similarities (
    id TEXT PRIMARY KEY,
    state_a_id TEXT NOT NULL,
    state_b_id TEXT NOT NULL,
    similarity REAL NOT NULL CHECK (similarity >= -1 AND similarity <= 1),
    computed_at TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE(state_a_id, state_b_id)
);

CREATE INDEX idx_similarities_state ON similarities(state_a_id);
CREATE INDEX idx_similarities_score ON similarities(similarity);

CREATE TABLE insights (
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
    dismissed_by TEXT,
    FOREIGN KEY (scope_id) REFERENCES entity_states(scope_id)
);

CREATE INDEX idx_insights_scope ON insights(scope_id);
CREATE INDEX idx_insights_active ON insights(scope_id, dismissed_at) WHERE dismissed_at IS NULL;

CREATE TABLE insight_cases (
    insight_id TEXT NOT NULL,
    entity_id TEXT NOT NULL,
    case_time TEXT NOT NULL,
    similarity REAL NOT NULL,
    outcome_diff REAL,
    FOREIGN KEY (insight_id) REFERENCES insights(id) ON DELETE CASCADE
);

CREATE INDEX idx_insight_cases ON insight_cases(insight_id);

CREATE TABLE insight_feedback (
    insight_id TEXT PRIMARY KEY,
    useful INTEGER, -- boolean: 0=false, 1=true
    applied INTEGER, -- boolean
    reason TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (insight_id) REFERENCES insights(id) ON DELETE CASCADE
);

-- Full-text search for insight text
CREATE VIRTUAL TABLE insights_fts USING fts5(title, summary, content='insights', content_rowid='rowid');

CREATE TRIGGER insights_fts_insert AFTER INSERT ON insights BEGIN
    INSERT INTO insights_fts(rowid, title, summary) VALUES (new.rowid, new.title, new.summary);
END;

CREATE TRIGGER insights_fts_delete AFTER DELETE ON insights BEGIN
    INSERT INTO insights_fts(insights_fts, rowid, title, summary) VALUES ('delete', old.rowid, old.title, old.summary);
END;

-- Schema version tracking
PRAGMA user_version = 1;
