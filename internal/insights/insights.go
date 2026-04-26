// Package insights provides generic pattern detection and insight generation.
package insights

import (
	"fmt"
	"sort"
	"time"

	"github.com/felixgeelhaar/chronos/internal/config"
	"github.com/felixgeelhaar/chronos/internal/similarity"
	"github.com/felixgeelhaar/chronos/pkg/insight"
	"github.com/felixgeelhaar/chronos/pkg/vector"
	"github.com/google/uuid"
)

// Generator creates insights by finding similar entity states and comparing outcomes.
type Generator struct {
	cfg *config.Config
}

// NewGenerator creates an insight generator with the given configuration.
func NewGenerator(cfg *config.Config) *Generator {
	return &Generator{cfg: cfg}
}

// Generate finds patterns in the provided entity states and returns insights.
// It compares each recent entity state against historical states within the same scope.
func (g *Generator) Generate(states []vector.EntityState) ([]insight.Insight, error) {
	if len(states) == 0 {
		return nil, nil
	}

	// Group states by scope
	byScope := make(map[uuid.UUID][]vector.EntityState)
	for _, s := range states {
		byScope[s.ScopeID] = append(byScope[s.ScopeID], s)
	}

	var allInsights []insight.Insight
	for scopeID, scopeStates := range byScope {
		insights := g.generateForScope(scopeID, scopeStates)
		allInsights = append(allInsights, insights...)
	}

	// Sort by confidence descending
	sort.Slice(allInsights, func(i, j int) bool {
		return allInsights[i].Confidence > allInsights[j].Confidence
	})

	// Limit per run
	if len(allInsights) > g.cfg.MaxInsightsPerRun {
		allInsights = allInsights[:g.cfg.MaxInsightsPerRun]
	}

	return allInsights, nil
}

func (g *Generator) generateForScope(scopeID uuid.UUID, states []vector.EntityState) []insight.Insight {
	if len(states) < 2 {
		return nil
	}

	// Find the most recent state for each entity
	latest := make(map[uuid.UUID]vector.EntityState)
	for _, s := range states {
		if existing, ok := latest[s.EntityID]; !ok || s.Timestamp.After(existing.Timestamp) {
			latest[s.EntityID] = s
		}
	}

	var insights []insight.Insight
	for entityID, recent := range latest {
		// Find similar historical states (excluding same entity's own history)
		var similar []insight.SimilarCase
		for _, candidate := range states {
			if candidate.EntityID == entityID {
				continue // Don't compare to self
			}
			if candidate.Timestamp.After(recent.Timestamp) || candidate.Timestamp.Equal(recent.Timestamp) {
				continue // Only look at historical data
			}

			sim := similarity.Cosine(recent.Features, candidate.Features)
			if sim >= g.cfg.SimilarityThreshold {
				// Compute outcome difference (last feature is assumed to be outcome metric)
				var outcomeDiff float64
				if len(recent.Features) > 0 && len(candidate.Features) > 0 {
					outcomeDiff = candidate.Features[len(candidate.Features)-1] - recent.Features[len(recent.Features)-1]
				}

				similar = append(similar, insight.SimilarCase{
					EntityID:    candidate.EntityID,
					Time:        candidate.Timestamp,
					Similarity:  sim,
					OutcomeDiff: outcomeDiff,
				})
			}
		}

		if len(similar) >= g.cfg.MinSampleSize {
			ins := g.buildInsight(scopeID, entityID, recent, similar)
			insights = append(insights, ins)
		}
	}

	return insights
}

func (g *Generator) buildInsight(scopeID, entityID uuid.UUID, recent vector.EntityState, similar []insight.SimilarCase) insight.Insight {
	// Sort by similarity
	sort.Slice(similar, func(i, j int) bool {
		return similar[i].Similarity > similar[j].Similarity
	})

	// Compute average outcome difference
	avgDiff := 0.0
	for _, s := range similar {
		avgDiff += s.OutcomeDiff
	}
	avgDiff /= float64(len(similar))

	// Confidence based on sample size and similarity consistency
	confidence := g.computeConfidence(similar)

	return insight.Insight{
		ID:            uuid.New(),
		ScopeID:       scopeID,
		Type:          "similar_state_pattern",
		SubjectEntity: entityID,
		SubjectTime:   recent.Timestamp,
		SimilarCases:  similar,
		SampleSize:    len(similar),
		Confidence:    confidence,
		Title:         fmt.Sprintf("Pattern detected for entity %s", entityID.String()[:8]),
		Summary:       fmt.Sprintf("Found %d similar historical cases with avg outcome difference %.2f", len(similar), avgDiff),
		Suggestion:    g.suggestionFromOutcome(avgDiff),
		GeneratedAt:   time.Now(),
	}
}

func (g *Generator) computeConfidence(similar []insight.SimilarCase) float64 {
	if len(similar) == 0 {
		return 0
	}

	// Base confidence from sample size (diminishing returns after 10)
	sampleFactor := 1.0
	if len(similar) < 5 {
		sampleFactor = float64(len(similar)) / 5.0
	}

	// Average similarity
	avgSim := 0.0
	for _, s := range similar {
		avgSim += s.Similarity
	}
	avgSim /= float64(len(similar))

	// Combine: high similarity + sufficient samples = high confidence
	confidence := avgSim * sampleFactor
	if confidence > 1.0 {
		confidence = 1.0
	}
	return confidence
}

func (g *Generator) suggestionFromOutcome(avgDiff float64) string {
	const threshold = 0.1
	if avgDiff < -threshold {
		return "Similar cases showed better outcomes. Consider reviewing current approach."
	}
	if avgDiff > threshold {
		return "Similar cases showed worse outcomes. Current approach may be favourable."
	}
	return "Outcomes were mixed among similar cases. Monitor closely."
}
