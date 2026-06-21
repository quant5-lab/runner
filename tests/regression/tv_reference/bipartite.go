package tv_reference

// maxBipartiteMatching finds a maximum-cardinality matching on a bipartite graph.
//
// adj[i] lists the right-side node indices reachable from left-side node i.
// Returns matchTV where matchTV[j] == i means right-node j is matched to
// left-node i; -1 means right-node j is unmatched.
func maxBipartiteMatching(adj [][]int, rightSize int) []int {
	matchTV := make([]int, rightSize)
	for j := range matchTV {
		matchTV[j] = -1
	}
	for i := range adj {
		augmentPath(adj, matchTV, i, make([]bool, rightSize))
	}
	return matchTV
}

// augmentPath extends the current matching by finding an augmenting path from
// left-node i via DFS. visited guards against revisiting right-side nodes within
// one augmentation attempt, ensuring the DFS terminates.
func augmentPath(adj [][]int, matchTV []int, i int, visited []bool) bool {
	for _, j := range adj[i] {
		if visited[j] {
			continue
		}
		visited[j] = true
		if matchTV[j] == -1 || augmentPath(adj, matchTV, matchTV[j], visited) {
			matchTV[j] = i
			return true
		}
	}
	return false
}
