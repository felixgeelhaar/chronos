package api

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// signalCursor is the opaque pagination token returned in the
// next_cursor field of a list response and accepted via the
// since_cursor query parameter. It is a (detected_at, id) tuple so
// callers can resume polling without depending on timestamp
// uniqueness — equal timestamps tie-break on the signal id.
type signalCursor struct {
	DetectedAt time.Time
	ID         uuid.UUID
}

// encodeSignalCursor renders a cursor to an opaque url-safe string.
// The format is intentionally compact (one base64 envelope around
// "<rfc3339nano>|<uuid>") so clients treat the token as opaque and
// don't construct it themselves.
func encodeSignalCursor(c signalCursor) string {
	raw := c.DetectedAt.UTC().Format(time.RFC3339Nano) + "|" + c.ID.String()
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

// decodeSignalCursor parses a token produced by encodeSignalCursor.
// Returns an error for any malformed input — the caller surfaces 400
// so a paste-error is loud rather than silently degrading into an
// unfiltered list.
func decodeSignalCursor(token string) (signalCursor, error) {
	if token == "" {
		return signalCursor{}, errors.New("empty cursor")
	}
	b, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return signalCursor{}, fmt.Errorf("decode cursor: %w", err)
	}
	parts := strings.SplitN(string(b), "|", 2)
	if len(parts) != 2 {
		return signalCursor{}, errors.New("cursor missing separator")
	}
	t, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return signalCursor{}, fmt.Errorf("cursor time: %w", err)
	}
	id, err := uuid.Parse(parts[1])
	if err != nil {
		return signalCursor{}, fmt.Errorf("cursor id: %w", err)
	}
	return signalCursor{DetectedAt: t, ID: id}, nil
}

// EncodeListCursor is the exported form of the HTTP/gRPC list cursor.
func EncodeListCursor(detectedAt time.Time, id uuid.UUID) string {
	return encodeSignalCursor(signalCursor{DetectedAt: detectedAt, ID: id})
}

// DecodeListCursor parses a token produced by EncodeListCursor.
func DecodeListCursor(token string) (time.Time, uuid.UUID, error) {
	c, err := decodeSignalCursor(token)
	if err != nil {
		return time.Time{}, uuid.Nil, err
	}
	return c.DetectedAt, c.ID, nil
}
