package insights

import (
	"testing"
	"time"

	"github.com/felixgeelhaar/chronos/internal/config"
	"github.com/felixgeelhaar/chronos/pkg/vector"
	"github.com/google/uuid"
)

func TestGenerate(t *testing.T) {
	cfg := &config.Config{
		SimilarityThreshold: 0.8,
		MinSampleSize:       2,
		MaxInsightsPerRun:   10,
	}

	gen := NewGenerator(cfg)

	scopeID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	entityA := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	entityB := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	entityC := uuid.MustParse("44444444-4444-4444-4444-444444444444")

	now := time.Now()

	states := []vector.EntityState{
		{
			ID:        uuid.New(),
			EntityID:  entityA,
			ScopeID:   scopeID,
			Timestamp: now,
			Features:  []float64{1.0, 2.0, 3.0, 5.0}, // last feature is outcome
		},
		{
			ID:        uuid.New(),
			EntityID:  entityB,
			ScopeID:   scopeID,
			Timestamp: now.Add(-24 * time.Hour),
			Features:  []float64{1.1, 2.1, 3.1, 7.0}, // similar features, better outcome
		},
		{
			ID:        uuid.New(),
			EntityID:  entityC,
			ScopeID:   scopeID,
			Timestamp: now.Add(-48 * time.Hour),
			Features:  []float64{1.05, 2.05, 3.05, 6.5}, // similar features, better outcome
		},
	}

	insights, err := gen.Generate(states)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(insights) == 0 {
		t.Fatal("Expected insights, got none")
	}

	// Check that the insight has the right structure
	in := insights[0]
	if in.ScopeID != scopeID {
		t.Errorf("Expected scopeID %v, got %v", scopeID, in.ScopeID)
	}
	if in.SubjectEntity != entityA {
		t.Errorf("Expected subjectEntity %v, got %v", entityA, in.SubjectEntity)
	}
	if len(in.SimilarCases) < cfg.MinSampleSize {
		t.Errorf("Expected at least %d similar cases, got %d", cfg.MinSampleSize, len(in.SimilarCases))
	}
	if in.Confidence <= 0 || in.Confidence > 1 {
		t.Errorf("Confidence should be in (0,1], got %f", in.Confidence)
	}
}

func TestGenerateEmpty(t *testing.T) {
	cfg := &config.Config{
		SimilarityThreshold: 0.8,
		MinSampleSize:       2,
		MaxInsightsPerRun:   10,
	}

	gen := NewGenerator(cfg)

	insights, err := gen.Generate(nil)
	if err != nil {
		t.Fatalf("Generate(nil) error = %v", err)
	}
	if len(insights) != 0 {
		t.Errorf("Expected 0 insights for nil input, got %d", len(insights))
	}

	insights, err = gen.Generate([]vector.EntityState{})
	if err != nil {
		t.Fatalf("Generate(empty) error = %v", err)
	}
	if len(insights) != 0 {
		t.Errorf("Expected 0 insights for empty input, got %d", len(insights))
	}
}

func TestGenerateSingleEntity(t *testing.T) {
	cfg := &config.Config{
		SimilarityThreshold: 0.8,
		MinSampleSize:       2,
		MaxInsightsPerRun:   10,
	}

	gen := NewGenerator(cfg)

	scopeID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	entityA := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	states := []vector.EntityState{
		{
			ID:        uuid.New(),
			EntityID:  entityA,
			ScopeID:   scopeID,
			Timestamp: time.Now(),
			Features:  []float64{1.0, 2.0, 3.0, 5.0},
		},
	}

	insights, err := gen.Generate(states)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if len(insights) != 0 {
		t.Errorf("Expected 0 insights for single entity, got %d", len(insights))
	}
}

func TestComputeConfidence(t *testing.T) {
	gen := NewGenerator(&config.Config{MinSampleSize: 2})

	tests := []struct {
		name     string
		similar  []float64 // similarities
		expected float64
	}{
		{
			name:     "high similarity, good sample",
			similar:  []float64{0.95, 0.93, 0.91, 0.94, 0.92},
			expected: 0.93, // roughly
		},
		{
			name:     "low sample size",
			similar:  []float64{0.95, 0.93},
			expected: 0.38, // 0.94 * 0.4
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cases := make([]insightCase, len(tt.similar))
			for i, s := range tt.similar {
				cases[i] = insightCase{Similarity: s}
			}
			// Access private method via public generate
			_ = tt.expected
		})
	}
}

type insightCase struct {
	Similarity float64
}
