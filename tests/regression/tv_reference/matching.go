package tv_reference

import (
	"strings"
	"time"
)

// MatchTolerance is the acceptable δ-window for pairing a runner entry to a TV entry.
type MatchTolerance struct {
	Time  time.Duration
	Price float64
}

// MatchPair records one runner↔TV pairing by index into their respective input slices.
type MatchPair struct {
	RunnerIdx int
	TVIdx     int
}

// closedRunnerPairs returns the maximum-cardinality set of runner↔TV entry pairs where:
//   - only closed runner trades (IsOpen == false) are eligible
//   - each trade participates in at most one pair
//   - entry time, price, and direction must agree within tol
//
// Indices in the returned pairs address the original (unsorted) slice positions.
// Maximum cardinality is guaranteed by bipartite matching (Kuhn's algorithm), so
// no valid pair is discarded due to greedy ordering.
func closedRunnerPairs(runner []RunnerTrade, tv []TVTrade, tol MatchTolerance) []MatchPair {
	closedIdx := closedRunnerIndices(runner)
	adj := buildCompatibilityGraph(closedIdx, runner, tv, tol)
	return pairsFromMatching(maxBipartiteMatching(adj, len(tv)), closedIdx)
}

func closedRunnerIndices(runner []RunnerTrade) []int {
	var idx []int
	for i, r := range runner {
		if !r.IsOpen {
			idx = append(idx, i)
		}
	}
	return idx
}

func buildCompatibilityGraph(closedIdx []int, runner []RunnerTrade, tv []TVTrade, tol MatchTolerance) [][]int {
	adj := make([][]int, len(closedIdx))
	for li, ri := range closedIdx {
		for j := range tv {
			if tradesCompatible(runner[ri], tv[j], tol) {
				adj[li] = append(adj[li], j)
			}
		}
	}
	return adj
}

func pairsFromMatching(matchTV []int, closedIdx []int) []MatchPair {
	var pairs []MatchPair
	for j, li := range matchTV {
		if li >= 0 {
			pairs = append(pairs, MatchPair{RunnerIdx: closedIdx[li], TVIdx: j})
		}
	}
	return pairs
}

func tradesCompatible(r RunnerTrade, t TVTrade, tol MatchTolerance) bool {
	return durationAbs(t.EntryUTC.Sub(r.EntryUTC)) <= tol.Time &&
		floatAbs(t.EntryPrice-r.EntryPrice) <= tol.Price &&
		directionsMatch(r.Direction, t.Direction)
}

func directionsMatch(runnerDir, tvDir string) bool {
	rn, rok := normalizeDirection(runnerDir)
	tn, tok := normalizeDirection(tvDir)
	return rok && tok && rn == tn
}

func normalizeDirection(s string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "long":
		return "long", true
	case "short":
		return "short", true
	default:
		return "", false
	}
}

func durationAbs(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}
	return d
}

func floatAbs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

func mathMax(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// ClosedRunnerPairs exposes closedRunnerPairs for callers that need pair indices, not just the aggregate counts returned by MatchExact.
func ClosedRunnerPairs(runner []RunnerTrade, tv []TVTrade, tol MatchTolerance) []MatchPair {
	return closedRunnerPairs(runner, tv, tol)
}
