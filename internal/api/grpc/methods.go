package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/felixgeelhaar/chronos"
	chronosv1 "github.com/felixgeelhaar/chronos/api/proto/chronos/v1"
	httapi "github.com/felixgeelhaar/chronos/internal/api"
	"github.com/felixgeelhaar/chronos/internal/ports"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// IngestBatch persists many observations in one RPC. Validation is
// all-or-nothing, matching POST /v1/ingest/batch.
func (s *Server) IngestBatch(ctx context.Context, req *chronosv1.IngestBatchRequest) (*chronosv1.IngestBatchResponse, error) {
	if s.states == nil {
		return nil, status.Error(codes.Unimplemented, "ingest disabled")
	}
	if req == nil || len(req.Observations) == 0 {
		return nil, status.Error(codes.InvalidArgument, "observations array is empty")
	}
	if len(req.Observations) > httapi.MaxIngestBatchSize {
		return nil, status.Error(codes.InvalidArgument,
			fmt.Sprintf("batch size %d exceeds max %d", len(req.Observations), httapi.MaxIngestBatchSize))
	}

	groups := make(map[string][]chronos.EntityState)
	for i, obs := range req.Observations {
		state, err := ingestRequestToEntityState(obs)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("observations[%d]: %v", i, err))
		}
		adapter := obs.Adapter
		if adapter == "" {
			adapter = "grpc"
		}
		groups[adapter] = append(groups[adapter], state)
	}

	accepted := 0
	for adapter, states := range groups {
		if err := s.states.Save(ctx, adapter, states); err != nil {
			s.logger.Error("grpc ingest batch failed", "adapter", adapter, "err", err)
			return nil, status.Error(codes.Internal, "ingest batch failed")
		}
		if s.metrics != nil {
			s.metrics.ObserveObservations(adapter, len(states))
		}
		accepted += len(states)
	}
	return &chronosv1.IngestBatchResponse{
		Accepted:       int32(accepted), //nolint:gosec // capped by MaxIngestBatchSize
		DeferDetection: req.DeferDetection,
	}, nil
}

// StreamSignals is the gRPC equivalent of GET /v1/signals/stream.
func (s *Server) StreamSignals(req *chronosv1.StreamSignalsRequest, stream chronosv1.ChronosService_StreamSignalsServer) error {
	if s.streamer == nil {
		return status.Error(codes.Unimplemented, "signal stream disabled (enable CHRONOS_DETECTION_INTERVAL)")
	}
	if req.ScopeId == "" && len(req.ScopeIds) == 0 {
		return status.Error(codes.InvalidArgument, "scope_id or scope_ids required")
	}
	if req.ScopeId != "" && len(req.ScopeIds) > 0 {
		return status.Error(codes.InvalidArgument, "set scope_id OR scope_ids, not both")
	}

	var scopes []uuid.UUID
	if req.ScopeId != "" {
		id, err := uuid.Parse(req.ScopeId)
		if err != nil {
			return status.Error(codes.InvalidArgument, "invalid scope_id")
		}
		scopes = []uuid.UUID{id}
	} else {
		scopes = make([]uuid.UUID, 0, len(req.ScopeIds))
		for _, raw := range req.ScopeIds {
			id, err := uuid.Parse(raw)
			if err != nil {
				return status.Error(codes.InvalidArgument, "invalid scope_ids entry")
			}
			scopes = append(scopes, id)
		}
	}

	pattern := ""
	if req.Pattern != chronosv1.PatternType_PATTERN_TYPE_UNSPECIFIED {
		pattern = string(patternTypeToDomain(req.Pattern))
	}

	id, ch := s.streamer.Subscribe(scopes, pattern)
	defer s.streamer.Unsubscribe(id)

	ctx := stream.Context()
	for {
		select {
		case <-ctx.Done():
			return nil
		case sig, ok := <-ch:
			if !ok {
				return nil
			}
			if err := stream.Send(fromDomainSignal(sig)); err != nil {
				return err
			}
		}
	}
}

// ValidateConfig dry-runs a candidate env-var map.
func (s *Server) ValidateConfig(_ context.Context, req *chronosv1.ValidateConfigRequest) (*chronosv1.ValidateConfigResponse, error) {
	env := map[string]string{}
	if req != nil {
		env = req.Env
	}
	cfg := httapi.LoadConfigWithOverrides(env)
	reports := httapi.BuildConfigReportForExport(cfg)
	out := make([]*chronosv1.DetectorReport, 0, len(reports))
	for _, r := range reports {
		th := make(map[string]string, len(r.Thresholds))
		for k, v := range r.Thresholds {
			th[k] = fmt.Sprint(v)
		}
		out = append(out, &chronosv1.DetectorReport{
			Name:       r.Name,
			Enabled:    r.Enabled,
			Reason:     r.Reason,
			Thresholds: th,
			Warnings:   r.Warnings,
		})
	}
	return &chronosv1.ValidateConfigResponse{Detectors: out}, nil
}

// ExportFederation returns anonymized pattern statistics.
func (s *Server) ExportFederation(ctx context.Context, _ *chronosv1.ExportFederationRequest) (*chronosv1.FederationExport, error) {
	if !httapi.FederationEnabled() {
		return nil, status.Error(codes.Unimplemented, "federation export disabled (set CHRONOS_FEDERATION_ENABLED=true to opt in)")
	}
	signals, err := s.signals.List(ctx, ports.SignalFilter{})
	if err != nil {
		s.logger.Error("grpc federation list failed", "err", err)
		return nil, status.Error(codes.Internal, "federation export failed")
	}
	stats := httapi.AggregateFederationStats(signals)
	patterns := make([]*chronosv1.FederationPatternStats, 0, len(stats))
	for _, row := range stats {
		patterns = append(patterns, &chronosv1.FederationPatternStats{
			Pattern:          row.Pattern,
			Count:            int32(row.Count), //nolint:gosec
			AvgStrength:      row.AvgStrength,
			MinStrength:      row.MinStrength,
			MaxStrength:      row.MaxStrength,
			AvgConfidence:    row.AvgConfidence,
			MinConfidence:    row.MinConfidence,
			MaxConfidence:    row.MaxConfidence,
			AvgSampleSize:    row.AvgSampleSize,
			TentativeCount:   int32(row.TentativeCount),   //nolint:gosec
			EstablishedCount: int32(row.EstablishedCount), //nolint:gosec
			StrongCount:      int32(row.StrongCount),      //nolint:gosec
		})
	}
	return &chronosv1.FederationExport{
		GeneratedAt:  timestamppb.New(time.Now().UTC()),
		Source:       "chronos",
		Version:      httapi.FederationExportVersion,
		Patterns:     patterns,
		TotalSignals: int32(len(signals)), //nolint:gosec
	}, nil
}
