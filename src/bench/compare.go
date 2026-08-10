// computes nullable statistics and selects benchmark comparison outcomes.
package bench

import (
	"math"
	"sort"
)

type NumberStatistics struct {
	Count             int      `json:"count"`
	Missing           int      `json:"missing"`
	Mean              *float64 `json:"mean"`
	Median            *float64 `json:"median"`
	Minimum           *float64 `json:"minimum"`
	Maximum           *float64 `json:"maximum"`
	StandardDeviation *float64 `json:"standardDeviation"`
}
type VariantStatistics struct {
	VariantID        string           `json:"variantId"`
	Count            int              `json:"count"`
	SuccessCount     int              `json:"successCount"`
	FailureCount     int              `json:"failureCount"`
	TimeoutCount     int              `json:"timeoutCount"`
	RequiredPassRate float64          `json:"requiredPassRate"`
	OptionalScore    NumberStatistics `json:"optionalScore"`
	DurationMS       NumberStatistics `json:"durationMs"`
	TotalTokens      NumberStatistics `json:"totalTokens"`
	EstimatedCost    NumberStatistics `json:"estimatedCost"`
	FilesChanged     NumberStatistics `json:"filesChanged"`
	Eligible         bool             `json:"eligible"`
}
type ComparisonResult struct {
	Outcome          string   `json:"outcome"`
	Winner           string   `json:"winner,omitempty"`
	Reason           string   `json:"reason"`
	Complete         bool     `json:"complete"`
	EligibleVariants []string `json:"eligibleVariants"`
}

func Summarize(results []RunResult, m Manifest) []VariantStatistics {
	out := make([]VariantStatistics, 0, len(m.Variants))
	for _, v := range m.Variants {
		var relevant []RunResult
		for _, r := range results {
			if r.VariantID == v.ID {
				relevant = append(relevant, r)
			}
		}
		s := VariantStatistics{VariantID: v.ID, Count: len(relevant)}
		var scores, durations, tokens, costs, filesChanged []float64
		missingScores, missingTokens, missingCosts := 0, 0, 0
		for _, r := range relevant {
			if r.Status == "completed" {
				s.SuccessCount++
			} else {
				s.FailureCount++
			}
			if r.Status == "timedOut" {
				s.TimeoutCount++
			}
			if r.RequiredPassed {
				s.RequiredPassRate++
			}
			durations = append(durations, float64(r.DurationMS))
			if r.OptionalScore != nil {
				scores = append(scores, *r.OptionalScore)
			} else {
				missingScores++
			}
			if r.Usage.TotalTokens != nil {
				tokens = append(tokens, float64(*r.Usage.TotalTokens))
			} else {
				missingTokens++
			}
			if r.Usage.EstimatedCost != nil {
				costs = append(costs, *r.Usage.EstimatedCost)
			} else {
				missingCosts++
			}
			filesChanged = append(filesChanged, float64(r.Repository.FilesCreated+r.Repository.FilesModified+r.Repository.FilesDeleted))
		}
		if s.Count > 0 {
			s.RequiredPassRate /= float64(s.Count)
		}
		s.OptionalScore = numberStats(scores, missingScores)
		s.DurationMS = numberStats(durations, 0)
		s.TotalTokens = numberStats(tokens, missingTokens)
		s.EstimatedCost = numberStats(costs, missingCosts)
		s.FilesChanged = numberStats(filesChanged, 0)
		s.Eligible = s.RequiredPassRate >= m.Comparison.MinimumRequiredPassRate
		out = append(out, s)
	}
	return out
}
func numberStats(values []float64, missing int) NumberStatistics {
	s := NumberStatistics{Count: len(values), Missing: missing}
	if len(values) == 0 {
		return s
	}
	sort.Float64s(values)
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))
	median := values[len(values)/2]
	if len(values)%2 == 0 {
		median = (values[len(values)/2-1] + median) / 2
	}
	min, max := values[0], values[len(values)-1]
	variance := 0.0
	for _, v := range values {
		variance += (v - mean) * (v - mean)
	}
	variance /= float64(len(values))
	std := math.Sqrt(variance)
	s.Mean = &mean
	s.Median = &median
	s.Minimum = &min
	s.Maximum = &max
	s.StandardDeviation = &std
	return s
}

func Compare(stats []VariantStatistics, complete bool, m Manifest) ComparisonResult {
	result := ComparisonResult{Complete: complete}
	if !complete {
		result.Outcome = "inconclusive"
		result.Reason = "run matrix is incomplete"
		return result
	}
	for _, s := range stats {
		if s.Eligible {
			result.EligibleVariants = append(result.EligibleVariants, s.VariantID)
		}
	}
	if len(stats) == 1 {
		if len(result.EligibleVariants) == 1 {
			result.Outcome = "evaluation-passed"
			result.Winner = stats[0].VariantID
			result.Reason = "all eligibility requirements passed"
		} else {
			result.Outcome = "evaluation-failed"
			result.Reason = "required pass rate is below the configured minimum"
		}
		return result
	}
	if len(result.EligibleVariants) == 0 {
		result.Outcome = "inconclusive"
		result.Reason = "no variant met the required pass-rate gate"
		return result
	}
	if len(result.EligibleVariants) == 1 {
		result.Outcome = "winner"
		result.Winner = result.EligibleVariants[0]
		result.Reason = "only this variant met the required pass-rate gate"
		return result
	}
	candidates := eligibleStats(stats)
	sort.SliceStable(candidates, func(i, j int) bool { return score(candidates[i]) > score(candidates[j]) })
	delta := score(candidates[0]) - score(candidates[1])
	if math.Abs(delta) > m.Comparison.TieThreshold {
		result.Outcome = "winner"
		result.Winner = candidates[0].VariantID
		result.Reason = "highest optional weighted score after required gating"
		return result
	}
	for _, tie := range m.Comparison.TieBreakers {
		winner := tieWinner(candidates, tie)
		if winner != "" {
			result.Outcome = "winner"
			result.Winner = winner
			result.Reason = "tie resolved by " + tie
			return result
		}
	}
	result.Outcome = "tie"
	result.Reason = "eligible variants remain within the tie threshold"
	return result
}
func eligibleStats(stats []VariantStatistics) []VariantStatistics {
	var out []VariantStatistics
	for _, s := range stats {
		if s.Eligible {
			out = append(out, s)
		}
	}
	return out
}
func score(s VariantStatistics) float64 {
	if s.OptionalScore.Mean != nil {
		return *s.OptionalScore.Mean
	}
	return s.RequiredPassRate
}
func tieWinner(stats []VariantStatistics, tie string) string {
	if len(stats) < 2 {
		return ""
	}
	a, b := stats[0], stats[1]
	switch tie {
	case "higherRequiredPassRate":
		if a.RequiredPassRate != b.RequiredPassRate {
			if a.RequiredPassRate > b.RequiredPassRate {
				return a.VariantID
			}
			return b.VariantID
		}
	case "higherMedianScore":
		if a.OptionalScore.Median != nil && b.OptionalScore.Median != nil && *a.OptionalScore.Median != *b.OptionalScore.Median {
			if *a.OptionalScore.Median > *b.OptionalScore.Median {
				return a.VariantID
			}
			return b.VariantID
		}
	case "lowerMedianDuration":
		if a.DurationMS.Median != nil && b.DurationMS.Median != nil && *a.DurationMS.Median != *b.DurationMS.Median {
			if *a.DurationMS.Median < *b.DurationMS.Median {
				return a.VariantID
			}
			return b.VariantID
		}
	case "lowerMedianTokens":
		if a.TotalTokens.Median != nil && b.TotalTokens.Median != nil && *a.TotalTokens.Median != *b.TotalTokens.Median {
			if *a.TotalTokens.Median < *b.TotalTokens.Median {
				return a.VariantID
			}
			return b.VariantID
		}
	case "fewerMedianFilesChanged":
		if a.FilesChanged.Median != nil && b.FilesChanged.Median != nil && *a.FilesChanged.Median != *b.FilesChanged.Median {
			if *a.FilesChanged.Median < *b.FilesChanged.Median {
				return a.VariantID
			}
			return b.VariantID
		}
	}
	return ""
}
