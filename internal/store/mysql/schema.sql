-- MySQL / MariaDB schema for Chronos.
--
-- Conventions:
--   * UUIDs are stored as CHAR(36) (MySQL has no native UUID type).
--   * Timestamps are DATETIME(6) so we keep microsecond precision; the
--     driver is configured with parseTime=true&loc=UTC to round-trip
--     time.Time correctly.
--   * Engine is implicit (InnoDB on every modern install) so we don't
--     pin it and break MariaDB / TiDB / PlanetScale, which all use
--     compatible defaults.
--   * Charset is implicit utf8mb4 — all string content is UTF-8.

CREATE TABLE IF NOT EXISTS entity_states (
    id          CHAR(36)     NOT NULL PRIMARY KEY,
    entity_id   CHAR(36)     NOT NULL,
    scope_id    CHAR(36)     NOT NULL,
    timestamp   DATETIME(6)  NOT NULL,
    features    JSON         NOT NULL,
    labels      JSON,
    meta        JSON,
    adapter     VARCHAR(255) NOT NULL,
    created_at  DATETIME(6)  NOT NULL,
    INDEX idx_entity_states_scope    (scope_id),
    INDEX idx_entity_states_entity   (entity_id),
    INDEX idx_entity_states_time     (timestamp),
    INDEX idx_entity_states_adapter  (adapter)
);

CREATE TABLE IF NOT EXISTS signals (
    id            CHAR(36)     NOT NULL PRIMARY KEY,
    scope_id      CHAR(36)     NOT NULL,
    series_id     CHAR(36)     NOT NULL,
    pattern       VARCHAR(64)  NOT NULL,
    detected_at   DATETIME(6)  NOT NULL,
    window_start  DATETIME(6)  NOT NULL,
    window_end    DATETIME(6)  NOT NULL,
    strength      DOUBLE       NOT NULL,
    confidence    DOUBLE       NOT NULL,
    metrics       JSON         NOT NULL,
    INDEX idx_signals_scope_time     (scope_id, detected_at),
    INDEX idx_signals_scope_pattern  (scope_id, pattern, detected_at),
    INDEX idx_signals_series         (series_id, detected_at)
);

CREATE TABLE IF NOT EXISTS signal_evidence (
    signal_id  CHAR(36)     NOT NULL,
    series_id  CHAR(36)     NOT NULL,
    time       DATETIME(6)  NOT NULL,
    kind       VARCHAR(64)  NOT NULL,
    score      DOUBLE       NOT NULL,
    metrics    JSON         NOT NULL,
    INDEX idx_signal_evidence (signal_id),
    CONSTRAINT fk_signal_evidence FOREIGN KEY (signal_id) REFERENCES signals(id) ON DELETE CASCADE
);
