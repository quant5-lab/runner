package tv_reference

import "testing"

func TestMaxBipartiteMatching_EmptyGraph(t *testing.T) {
	matchTV := maxBipartiteMatching(nil, 0)
	if len(matchTV) != 0 {
		t.Errorf("len(matchTV) = %d, want 0", len(matchTV))
	}
}

func TestMaxBipartiteMatching_NoEdges(t *testing.T) {
	adj := [][]int{{}, {}, {}}
	matchTV := maxBipartiteMatching(adj, 3)
	for j, li := range matchTV {
		if li != -1 {
			t.Errorf("matchTV[%d] = %d, want -1 (no edges)", j, li)
		}
	}
}

func TestMaxBipartiteMatching_PerfectOneToOne(t *testing.T) {
	adj := [][]int{{0}, {1}, {2}}
	matchTV := maxBipartiteMatching(adj, 3)
	matched := 0
	for _, li := range matchTV {
		if li >= 0 {
			matched++
		}
	}
	if matched != 3 {
		t.Errorf("matched = %d, want 3", matched)
	}
}

func TestMaxBipartiteMatching_MoreLeftThanRight(t *testing.T) {
	adj := [][]int{{0, 1}, {0, 1}, {0, 1}, {0, 1}}
	matchTV := maxBipartiteMatching(adj, 2)
	matched := 0
	for _, li := range matchTV {
		if li >= 0 {
			matched++
		}
	}
	if matched != 2 {
		t.Errorf("matched = %d, want 2 (right-side limited)", matched)
	}
}

// TestMaxBipartiteMatching_AugmentingPathRequired exercises the case where greedy
// assignment of the first left-node blocks the second, but an augmenting path
// re-routes the first so both are matched.
//
// Graph:  L0 → {R0, R1}
//
//	L1 → {R0}
//
// Greedy (L0 first): L0→R0, L1 stranded → matched=1.
// Augmenting path:   L0→R0 initially; augmenting L1 re-routes L0 to R1 → matched=2.
func TestMaxBipartiteMatching_AugmentingPathRequired(t *testing.T) {
	adj := [][]int{
		{0, 1}, // L0 compat R0 and R1
		{0},    // L1 compat R0 only
	}
	matchTV := maxBipartiteMatching(adj, 2)
	matched := 0
	for _, li := range matchTV {
		if li >= 0 {
			matched++
		}
	}
	if matched != 2 {
		t.Errorf("matched = %d, want 2 (augmenting path must re-route L0 to R1)", matched)
	}
}

func TestMaxBipartiteMatching_InputOrderInvariant(t *testing.T) {
	adj1 := [][]int{{0, 1}, {0}}
	adj2 := [][]int{{0}, {0, 1}}

	m1 := countMatched(maxBipartiteMatching(adj1, 2))
	m2 := countMatched(maxBipartiteMatching(adj2, 2))
	if m1 != m2 {
		t.Errorf("matched counts differ by input order: %d vs %d", m1, m2)
	}
}

// TestMaxBipartiteMatching_ZeroLeftNonzeroRight confirms that when there are no
// left-side nodes, all right-side slots remain unmatched (-1) and the result slice
// length equals rightSize.
func TestMaxBipartiteMatching_ZeroLeftNonzeroRight(t *testing.T) {
	matchTV := maxBipartiteMatching(nil, 5)
	if len(matchTV) != 5 {
		t.Fatalf("len(matchTV) = %d, want 5", len(matchTV))
	}
	for j, li := range matchTV {
		if li != -1 {
			t.Errorf("matchTV[%d] = %d, want -1 (no left nodes)", j, li)
		}
	}
}

// TestMaxBipartiteMatching_DeepAugmentingChain verifies that the algorithm
// correctly resolves a chain where achieving maximum cardinality requires
// re-routing four already-matched left-nodes in sequence (augmenting path depth 4).
//
// Graph (each → lists right-side indices in adjacency order, i.e. preferred slot first):
//
//	L0 → {R0, R4}   grabs R0 first; must be re-routed to its escape R4
//	L1 → {R1, R0}   grabs R1 first; displaced to R0 when L2 needs R1
//	L2 → {R2, R1}   grabs R2 first; displaced to R1 when L3 needs R2
//	L3 → {R3, R2}   grabs R3 first; displaced to R2 when L4 needs R3
//	L4 → {R3}       single option; triggers the whole 4-hop chain
//
// Initial greedy assignment: L0→R0, L1→R1, L2→R2, L3→R3.
// L4 forces: R3←L4, L3→R2, L2→R1, L1→R0, L0→R4 — a depth-4 augmenting path.
func TestMaxBipartiteMatching_DeepAugmentingChain(t *testing.T) {
	adj := [][]int{
		{0, 4},
		{1, 0},
		{2, 1},
		{3, 2},
		{3},
	}
	matchTV := maxBipartiteMatching(adj, 5)
	if got := countMatched(matchTV); got != 5 {
		t.Errorf("matched = %d, want 5 (deep chain must re-route L0 through L3)", got)
	}
}

func countMatched(matchTV []int) int {
	n := 0
	for _, li := range matchTV {
		if li >= 0 {
			n++
		}
	}
	return n
}
