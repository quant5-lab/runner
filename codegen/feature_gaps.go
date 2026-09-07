package codegen

func deduplicateFeatureGaps(gaps []string) []string {
	if len(gaps) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(gaps))
	out := make([]string, 0, len(gaps))
	for _, g := range gaps {
		if _, exists := seen[g]; !exists {
			seen[g] = struct{}{}
			out = append(out, g)
		}
	}
	return out
}
