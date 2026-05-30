package preprocessor

// NormalizeIfBlocks is a no-op passthrough — the Pine Script parser handles
// indented if/else blocks natively without requiring text-level normalization.
func NormalizeIfBlocks(script string) string {
	return script
}
