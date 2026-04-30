package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	chronosv1 "github.com/felixgeelhaar/chronos/api/proto/chronos/v1"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// GRPCClient is a gRPC transport implementation of the Chronos client
// surface. It returns the same Signal and IngestRequest types as the HTTP
// [Client] so callers can switch transports without changing business code.
type GRPCClient struct {
	conn   *grpc.ClientConn
	client chronosv1.ChronosServiceClient
	token  string
}

// GRPCOption configures a GRPCClient.
type GRPCOption func(*GRPCClient)

// WithGRPCToken sets a bearer token sent as gRPC metadata on every RPC.
func WithGRPCToken(tok string) GRPCOption { return func(c *GRPCClient) { c.token = tok } }

// NewGRPC constructs a GRPCClient targeting addr (e.g. "localhost:7779").
func NewGRPC(addr string, opts ...GRPCOption) (*GRPCClient, error) {
	if addr == "" {
		return nil, errors.New("chronos grpc client: address required")
	}
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(16<<20)), // 16 MiB
	)
	if err != nil {
		return nil, fmt.Errorf("chronos grpc client: dial: %w", err)
	}
	c := &GRPCClient{
		conn:   conn,
		client: chronosv1.NewChronosServiceClient(conn),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c, nil
}

// Close tears down the underlying gRPC connection.
func (c *GRPCClient) Close() error { return c.conn.Close() }

// Ingest sends a single observation to the gRPC Ingest RPC.
func (c *GRPCClient) Ingest(ctx context.Context, req IngestRequest) (uuid.UUID, error) {
	protoReq := &chronosv1.IngestRequest{
		Id:       req.ID.String(),
		EntityId: req.EntityID.String(),
		ScopeId:  req.ScopeID.String(),
		Features: req.Features,
		Labels:   req.Labels,
		Meta:     req.Meta,
		Adapter:  req.Adapter,
	}
	if !req.Timestamp.IsZero() {
		protoReq.Timestamp = timestamppb.New(req.Timestamp)
	}

	resp, err := c.client.Ingest(c.withAuth(ctx), protoReq)
	if err != nil {
		return uuid.Nil, fmt.Errorf("chronos grpc client: ingest: %w", err)
	}
	id, err := uuid.Parse(resp.Id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("chronos grpc client: parse response id: %w", err)
	}
	return id, nil
}

// Signals returns a fluent builder for the ListSignals RPC.
func (c *GRPCClient) Signals() *GRPCSignalQuery {
	return &GRPCSignalQuery{c: c}
}

// GRPCSignalQuery is the gRPC equivalent of SignalQuery.
type GRPCSignalQuery struct {
	c             *GRPCClient
	scope         uuid.UUID
	series        *uuid.UUID
	pattern       *string
	since, until  *time.Time
	minConfidence *float64
	limit         int
}

// Scope filters by scope ID. Required for List.
func (q *GRPCSignalQuery) Scope(id uuid.UUID) *GRPCSignalQuery {
	q.scope = id
	return q
}

// Series filters to signals about a single entity.
func (q *GRPCSignalQuery) Series(id uuid.UUID) *GRPCSignalQuery {
	q.series = &id
	return q
}

// Pattern filters to a single PatternType (use the PatternType* constants).
func (q *GRPCSignalQuery) Pattern(name string) *GRPCSignalQuery {
	q.pattern = &name
	return q
}

// Since restricts results to signals detected at or after t.
func (q *GRPCSignalQuery) Since(t time.Time) *GRPCSignalQuery {
	q.since = &t
	return q
}

// Until restricts results to signals detected before t.
func (q *GRPCSignalQuery) Until(t time.Time) *GRPCSignalQuery {
	q.until = &t
	return q
}

// MinConfidence drops signals below threshold.
func (q *GRPCSignalQuery) MinConfidence(threshold float64) *GRPCSignalQuery {
	q.minConfidence = &threshold
	return q
}

// Limit caps the number of returned signals.
func (q *GRPCSignalQuery) Limit(n int) *GRPCSignalQuery {
	q.limit = n
	return q
}

// List runs the query via the ListSignals RPC.
func (q *GRPCSignalQuery) List(ctx context.Context) ([]Signal, error) {
	if q.scope == uuid.Nil {
		return nil, errors.New("chronos grpc client: Scope is required")
	}
	req := &chronosv1.ListSignalsRequest{
		ScopeId: q.scope.String(),
		Limit:   int32(q.limit),
	}
	if q.series != nil {
		req.Series = q.series.String()
	}
	if q.pattern != nil {
		req.Pattern = patternTypeFromString(*q.pattern)
	}
	if q.since != nil {
		req.Since = timestamppb.New(*q.since)
	}
	if q.until != nil {
		req.Until = timestamppb.New(*q.until)
	}
	if q.minConfidence != nil {
		req.MinConfidence = *q.minConfidence
	}

	resp, err := q.c.client.ListSignals(q.c.withAuth(ctx), req)
	if err != nil {
		return nil, fmt.Errorf("chronos grpc client: list signals: %w", err)
	}

	signals := make([]Signal, 0, len(resp.Signals))
	for _, ps := range resp.Signals {
		signals = append(signals, protoToSignal(ps))
	}
	return signals, nil
}

// Get returns a single signal by ID via the GetSignal RPC.
func (q *GRPCSignalQuery) Get(ctx context.Context, id uuid.UUID) (Signal, error) {
	resp, err := q.c.client.GetSignal(q.c.withAuth(ctx), &chronosv1.GetSignalRequest{Id: id.String()})
	if err != nil {
		return Signal{}, fmt.Errorf("chronos grpc client: get signal: %w", err)
	}
	return protoToSignal(resp), nil
}

// withAuth injects bearer token metadata when a token is configured.
func (c *GRPCClient) withAuth(ctx context.Context) context.Context {
	if c.token == "" {
		return ctx
	}
	return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+c.token)
}

// patternTypeFromString maps a client PatternType constant to the proto enum.
func patternTypeFromString(s string) chronosv1.PatternType {
	switch s {
	case PatternTypeRecurrence:
		return chronosv1.PatternType_PATTERN_TYPE_RECURRENCE
	case PatternTypeTrend:
		return chronosv1.PatternType_PATTERN_TYPE_TREND
	case PatternTypeSpike:
		return chronosv1.PatternType_PATTERN_TYPE_SPIKE
	case PatternTypeDrop:
		return chronosv1.PatternType_PATTERN_TYPE_DROP
	case PatternTypeStall:
		return chronosv1.PatternType_PATTERN_TYPE_STALL
	case PatternTypeAnomaly:
		return chronosv1.PatternType_PATTERN_TYPE_ANOMALY
	case PatternTypeSeasonality:
		return chronosv1.PatternType_PATTERN_TYPE_SEASONALITY
	case PatternTypeCorrelation:
		return chronosv1.PatternType_PATTERN_TYPE_CORRELATION
	default:
		return chronosv1.PatternType_PATTERN_TYPE_UNSPECIFIED
	}
}

// patternTypeToString maps a proto PatternType to the client PatternType constant.
func patternTypeToString(p chronosv1.PatternType) string {
	switch p {
	case chronosv1.PatternType_PATTERN_TYPE_RECURRENCE:
		return PatternTypeRecurrence
	case chronosv1.PatternType_PATTERN_TYPE_TREND:
		return PatternTypeTrend
	case chronosv1.PatternType_PATTERN_TYPE_SPIKE:
		return PatternTypeSpike
	case chronosv1.PatternType_PATTERN_TYPE_DROP:
		return PatternTypeDrop
	case chronosv1.PatternType_PATTERN_TYPE_STALL:
		return PatternTypeStall
	case chronosv1.PatternType_PATTERN_TYPE_ANOMALY:
		return PatternTypeAnomaly
	case chronosv1.PatternType_PATTERN_TYPE_SEASONALITY:
		return PatternTypeSeasonality
	case chronosv1.PatternType_PATTERN_TYPE_CORRELATION:
		return PatternTypeCorrelation
	default:
		return ""
	}
}

// protoToSignal converts a proto Signal to the client.Signal wire type.
func protoToSignal(ps *chronosv1.Signal) Signal {
	if ps == nil {
		return Signal{}
	}
	evidence := make([]Evidence, 0, len(ps.Evidence))
	for _, pe := range ps.Evidence {
		if pe == nil {
			continue
		}
		evidence = append(evidence, Evidence{
			Series:  parseUUID(pe.Series),
			Time:    pe.Time.AsTime(),
			Kind:    pe.Kind,
			Score:   pe.Score,
			Metrics: pe.Metrics,
		})
	}
	return Signal{
		ID:         parseUUID(ps.Id),
		ScopeID:    parseUUID(ps.ScopeId),
		Series:     parseUUID(ps.Series),
		Pattern:    patternTypeToString(ps.Pattern),
		DetectedAt: ps.DetectedAt.AsTime(),
		Window: TimeWindow{
			Start: ps.Window.Start.AsTime(),
			End:   ps.Window.End.AsTime(),
		},
		Strength:   ps.Strength,
		Confidence: ps.Confidence,
		Metrics:    ps.Metrics,
		Evidence:   evidence,
	}
}

// parseUUID safely parses a UUID string; returns uuid.Nil on error.
func parseUUID(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}
