package detect

import "github.com/google/uuid"

// chronosCrossScopeNamespace is the fixed UUIDv5 namespace used to
// hash scope/series identifiers when AnonymizeCrossScope is on. Keeping
// it constant across the codebase means the same input always hashes
// to the same opaque uuid — downstream consumers can still group
// signals that involve the same anonymous entity across runs without
// learning who that entity is.
var chronosCrossScopeNamespace = uuid.Must(uuid.Parse("c5a5c0fe-c30c-4c0e-8c0f-c50fec5050fe"))

// anonymizeID returns a deterministic UUIDv5 derived from id under the
// cross-scope namespace. The output is not reversible (SHA-1 of the
// raw bytes); two callers passing the same id get the same opaque
// uuid, two different ids get different opaque uuids with
// overwhelmingly high probability.
func anonymizeID(id uuid.UUID) uuid.UUID {
	return uuid.NewSHA1(chronosCrossScopeNamespace, id[:])
}
