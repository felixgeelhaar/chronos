CREATE TABLE IF NOT EXISTS entity_states (
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
CREATE INDEX IF NOT EXISTS idx_entity_states_scope   ON entity_states(scope_id);
CREATE INDEX IF NOT EXISTS idx_entity_states_entity  ON entity_states(entity_id);
CREATE INDEX IF NOT EXISTS idx_entity_states_time    ON entity_states(timestamp);
CREATE INDEX IF NOT EXISTS idx_entity_states_adapter ON entity_states(adapter);

CREATE TABLE IF NOT EXISTS signals (
    id UUID PRIMARY KEY,
    scope_id UUID NOT NULL,
    series_id UUID NOT NULL,
    pattern TEXT NOT NULL,
    detected_at TIMESTAMPTZ NOT NULL,
    window_start TIMESTAMPTZ NOT NULL,
    window_end TIMESTAMPTZ NOT NULL,
    strength DOUBLE PRECISION NOT NULL CHECK (strength >= 0 AND strength <= 1),
    confidence DOUBLE PRECISION NOT NULL CHECK (confidence >= 0 AND confidence <= 1),
    metrics JSONB NOT NULL DEFAULT '{}'::jsonb
);
CREATE INDEX IF NOT EXISTS idx_signals_scope_time    ON signals(scope_id, detected_at DESC);
CREATE INDEX IF NOT EXISTS idx_signals_scope_pattern ON signals(scope_id, pattern, detected_at DESC);
CREATE INDEX IF NOT EXISTS idx_signals_series        ON signals(series_id, detected_at DESC);

CREATE TABLE IF NOT EXISTS signal_evidence (
    signal_id UUID NOT NULL REFERENCES signals(id) ON DELETE CASCADE,
    series_id UUID NOT NULL,
    time TIMESTAMPTZ NOT NULL,
    kind TEXT NOT NULL,
    score DOUBLE PRECISION NOT NULL,
    metrics JSONB NOT NULL DEFAULT '{}'::jsonb
);
CREATE INDEX IF NOT EXISTS idx_signal_evidence ON signal_evidence(signal_id);
