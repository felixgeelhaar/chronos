package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"sync"

	"github.com/felixgeelhaar/chronos/internal/config"
)

// ConfigValidateRequest is the wire shape for POST /v1/config/validate.
// Callers pass the env vars they intend to flip; the handler reports
// what each detector would do under that config WITHOUT mutating the
// running server's behaviour.
type ConfigValidateRequest struct {
	Env map[string]string `json:"env"`
}

// DetectorReport carries the dry-run verdict for one detector.
type DetectorReport struct {
	Name       string         `json:"name"`
	Enabled    bool           `json:"enabled"`
	Reason     string         `json:"reason,omitempty"`
	Thresholds map[string]any `json:"thresholds,omitempty"`
	Warnings   []string       `json:"warnings,omitempty"`
}

// ConfigValidateResponse lists one report per detector, sorted by
// name for stable diffing across runs.
type ConfigValidateResponse struct {
	Detectors []DetectorReport `json:"detectors"`
}

// configValidateMu serialises the env-var swap inside the handler so
// concurrent validate requests cannot race against each other or
// against the live config reader. The endpoint is ops-only and rare;
// the mutex's contention cost is irrelevant.
var configValidateMu sync.Mutex

// handleConfigValidate dry-runs a candidate config and reports the
// per-detector outcome. It applies overrides in-process via Setenv +
// restore (guarded by a mutex), so callers see the same Default()
// loader path that the live server uses — no parallel parser to
// drift apart from production.
func (s *Server) handleConfigValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req ConfigValidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	cfg := loadConfigWithOverrides(req.Env)
	respondJSON(w, http.StatusOK, buildConfigReport(cfg))
}

// loadConfigWithOverrides snapshots the touched env vars, applies the
// overrides, calls Default(), then restores. The mutex makes the swap
// atomic with respect to other validate calls.
func loadConfigWithOverrides(env map[string]string) *config.Config {
	configValidateMu.Lock()
	defer configValidateMu.Unlock()

	type snap struct{ value string; present bool }
	saved := make(map[string]snap, len(env))
	for k, v := range env {
		prev, ok := os.LookupEnv(k)
		saved[k] = snap{value: prev, present: ok}
		_ = os.Setenv(k, v)
	}
	defer func() {
		for k, s := range saved {
			if s.present {
				_ = os.Setenv(k, s.value)
			} else {
				_ = os.Unsetenv(k)
			}
		}
	}()
	return config.Default()
}

// buildConfigReport classifies each detector under the candidate cfg.
// "Enabled" means the detector's gating thresholds will let at least
// some inputs through; "disabled" reports the specific gating value
// that suppresses the detector so an operator sees exactly which knob
// to turn.
func buildConfigReport(cfg *config.Config) ConfigValidateResponse {
	reports := []DetectorReport{
		trendReport(cfg),
		spikeReport(cfg),
		dropReport(cfg),
		stallReport(cfg),
		anomalyReport(cfg),
		seasonalityReport(cfg),
		correlationReport(cfg),
		changePointReport(cfg),
		outlierClusterReport(cfg),
		crossScopeCorrelationReport(cfg),
	}
	sort.Slice(reports, func(i, j int) bool { return reports[i].Name < reports[j].Name })
	return ConfigValidateResponse{Detectors: reports}
}

func trendReport(cfg *config.Config) DetectorReport {
	r := DetectorReport{
		Name:       "trend",
		Thresholds: map[string]any{"min_slope": cfg.TrendMinSlope, "min_points": cfg.TrendMinPoints},
	}
	switch {
	case cfg.TrendMinPoints < 2:
		r.Reason = fmt.Sprintf("min_points %d < 2 disables this detector", cfg.TrendMinPoints)
	case cfg.TrendMinSlope <= 0:
		r.Reason = fmt.Sprintf("min_slope %v ≤ 0 disables this detector", cfg.TrendMinSlope)
	default:
		r.Enabled = true
	}
	return r
}

func spikeReport(cfg *config.Config) DetectorReport {
	r := DetectorReport{
		Name:       "spike",
		Thresholds: map[string]any{"z_score": cfg.SpikeZScore, "window": cfg.SpikeWindow},
	}
	switch {
	case cfg.SpikeWindow < 2:
		r.Reason = fmt.Sprintf("window %d < 2 disables this detector", cfg.SpikeWindow)
	case cfg.SpikeZScore <= 0:
		r.Reason = fmt.Sprintf("z_score %v ≤ 0 disables this detector", cfg.SpikeZScore)
	default:
		r.Enabled = true
	}
	return r
}

func dropReport(cfg *config.Config) DetectorReport {
	r := DetectorReport{
		Name:       "drop",
		Thresholds: map[string]any{"z_score": cfg.DropZScore, "window": cfg.SpikeWindow},
	}
	switch {
	case cfg.SpikeWindow < 2:
		r.Reason = fmt.Sprintf("window %d < 2 disables this detector", cfg.SpikeWindow)
	case cfg.DropZScore <= 0:
		r.Reason = fmt.Sprintf("z_score %v ≤ 0 disables this detector", cfg.DropZScore)
	default:
		r.Enabled = true
	}
	return r
}

func stallReport(cfg *config.Config) DetectorReport {
	r := DetectorReport{
		Name:       "stall",
		Thresholds: map[string]any{"max_stddev": cfg.StallMaxStdDev, "min_points": cfg.StallMinPoints},
	}
	switch {
	case cfg.StallMinPoints < 2:
		r.Reason = fmt.Sprintf("min_points %d < 2 disables this detector", cfg.StallMinPoints)
	case cfg.StallMaxStdDev <= 0:
		r.Reason = fmt.Sprintf("max_stddev %v ≤ 0 disables this detector", cfg.StallMaxStdDev)
	default:
		r.Enabled = true
	}
	return r
}

func anomalyReport(cfg *config.Config) DetectorReport {
	r := DetectorReport{
		Name:       "anomaly",
		Thresholds: map[string]any{"max_similarity": cfg.AnomalyMaxSimilarity, "min_peers": cfg.AnomalyMinPeers},
	}
	switch {
	case cfg.AnomalyMinPeers < 1:
		r.Reason = fmt.Sprintf("min_peers %d < 1 disables this detector", cfg.AnomalyMinPeers)
	case cfg.AnomalyMaxSimilarity <= 0:
		r.Reason = fmt.Sprintf("max_similarity %v ≤ 0 disables this detector", cfg.AnomalyMaxSimilarity)
	default:
		r.Enabled = true
	}
	return r
}

func seasonalityReport(cfg *config.Config) DetectorReport {
	r := DetectorReport{
		Name: "seasonality",
		Thresholds: map[string]any{
			"min_autocorr": cfg.SeasonalityMinAutocorr,
			"min_points":   cfg.SeasonalityMinPoints,
			"min_period":   cfg.SeasonalityMinPeriod,
		},
	}
	switch {
	case cfg.SeasonalityMinPoints < cfg.SeasonalityMinPeriod*2:
		r.Reason = fmt.Sprintf("min_points %d < 2×min_period (%d) — not enough room to fit two cycles",
			cfg.SeasonalityMinPoints, cfg.SeasonalityMinPeriod*2)
	case cfg.SeasonalityMinAutocorr <= 0:
		r.Reason = fmt.Sprintf("min_autocorr %v ≤ 0 disables this detector", cfg.SeasonalityMinAutocorr)
	default:
		r.Enabled = true
	}
	return r
}

func correlationReport(cfg *config.Config) DetectorReport {
	r := DetectorReport{
		Name:       "correlation",
		Thresholds: map[string]any{"min_r": cfg.CorrelationMin, "min_points": cfg.CorrelationMinPoints},
	}
	switch {
	case cfg.CorrelationMinPoints < 2:
		r.Reason = fmt.Sprintf("min_points %d < 2 disables this detector", cfg.CorrelationMinPoints)
	case cfg.CorrelationMin <= 0:
		r.Reason = fmt.Sprintf("min_r %v ≤ 0 disables this detector", cfg.CorrelationMin)
	default:
		r.Enabled = true
	}
	return r
}

func changePointReport(cfg *config.Config) DetectorReport {
	r := DetectorReport{
		Name:       "change_point",
		Thresholds: map[string]any{"min_shift": cfg.ChangePointMinShift, "min_points": cfg.ChangePointMinPoints},
	}
	switch {
	case cfg.ChangePointMinPoints < 4:
		r.Reason = fmt.Sprintf("min_points %d < 4 — need ≥ 2 either side of the split", cfg.ChangePointMinPoints)
	case cfg.ChangePointMinShift <= 0:
		r.Reason = fmt.Sprintf("min_shift %v ≤ 0 disables this detector", cfg.ChangePointMinShift)
	default:
		r.Enabled = true
	}
	return r
}

func outlierClusterReport(cfg *config.Config) DetectorReport {
	r := DetectorReport{
		Name: "outlier_cluster",
		Thresholds: map[string]any{
			"min_series":  cfg.OutlierClusterMinSeries,
			"z":           cfg.OutlierClusterZ,
			"window_secs": cfg.OutlierClusterTimeWindow.Seconds(),
		},
	}
	switch {
	case cfg.OutlierClusterMinSeries < 2:
		r.Reason = fmt.Sprintf("min_series %d < 2 disables this detector", cfg.OutlierClusterMinSeries)
	case cfg.OutlierClusterZ <= 0:
		r.Reason = fmt.Sprintf("z %v ≤ 0 disables this detector", cfg.OutlierClusterZ)
	default:
		r.Enabled = true
	}
	return r
}

func crossScopeCorrelationReport(cfg *config.Config) DetectorReport {
	r := DetectorReport{
		Name: "cross_scope_correlation",
		Thresholds: map[string]any{
			"min_r":      cfg.CrossScopeMin,
			"min_points": cfg.CrossScopeMinPoints,
			"anonymize":  cfg.AnonymizeCrossScope,
		},
	}
	switch {
	case cfg.CrossScopeMinPoints < 2:
		r.Reason = fmt.Sprintf("min_points %d < 2 disables this detector", cfg.CrossScopeMinPoints)
	case cfg.CrossScopeMin <= 0:
		r.Reason = fmt.Sprintf("min_r %v ≤ 0 disables this detector", cfg.CrossScopeMin)
	default:
		r.Enabled = true
		if !cfg.AnonymizeCrossScope {
			r.Warnings = append(r.Warnings,
				"AnonymizeCrossScope is off — emitted signals will carry real tenant scope_ids (see #20)")
		}
	}
	return r
}
