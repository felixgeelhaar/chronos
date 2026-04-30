
## gRPC transport

Add gRPC as a first-class transport alongside the existing HTTP REST API. Define protobuf messages for EntityState, Signal, and service methods for Ingest and ListSignals. Implement a gRPC server in internal/api/grpc/, wire it into cmd/chronos/serve.go, and provide a gRPC client constructor in client/. Both transports share the same domain ports (EntityStateRepository, SignalRepository). The gRPC service mirrors the HTTP surface: Ingest (streaming or unary) and ListSignals (with ScopeID, PatternType, MinConfidence filters). Include reflection and health checks for operational discoverability.

---
