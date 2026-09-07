// Package featuregap records unimplemented Pine builtin calls encountered at runtime.
// The collector is process-global and deduplicates warnings per (name, source) pair;
// each generated binary is a single process so state never leaks across strategies.
package featuregap

import (
	"fmt"
	"log"
	"math"
	"sort"
	"sync"
)

type Hit struct {
	Name         string   `json:"featureId"`
	Source       string   `json:"source"`
	File         string   `json:"file,omitempty"`
	Line         int      `json:"line,omitempty"`
	Column       int      `json:"column,omitempty"`
	Phase        string   `json:"phase"`
	Impact       string   `json:"impact"`
	Substitution string   `json:"substitution"`
	Sinks        []string `json:"affectedSinks,omitempty"`
	FirstSeenBar int      `json:"firstReachedBar"`
	LastSeenBar  int      `json:"lastReachedBar"`
	HitCount     int      `json:"hitCount"`
}

type key struct {
	name   string
	source string
}

var (
	mu     sync.Mutex
	hits   = map[key]*Hit{}
	warned = map[key]struct{}{}
)

func Record(name, source string, barIdx int) float64 {
	return RecordDiagnostic(name, source, "runtime", "observable-non-backtest", "NaN", nil, barIdx)
}

func RecordDiagnostic(name, source, phase, impact, substitution string, sinks []string, barIdx int) float64 {
	return RecordDiagnosticAt(name, source, "", 0, 0, phase, impact, substitution, sinks, barIdx)
}

func RecordDiagnosticAt(name, source, file string, line, column int, phase, impact, substitution string, sinks []string, barIdx int) float64 {
	mu.Lock()
	k := key{name: name, source: source}
	h, ok := hits[k]
	if !ok {
		h = &Hit{
			Name:         name,
			Source:       source,
			File:         file,
			Line:         line,
			Column:       column,
			Phase:        phase,
			Impact:       impact,
			Substitution: substitution,
			Sinks:        append([]string(nil), sinks...),
			FirstSeenBar: barIdx,
			LastSeenBar:  barIdx,
		}
		hits[k] = h
	}
	if impactRank(impact) > impactRank(h.Impact) {
		h.Impact = impact
	}
	if h.File == "" && file != "" {
		h.File = file
		h.Line = line
		h.Column = column
	}
	h.Sinks = mergeStrings(h.Sinks, sinks)
	h.HitCount++
	if h.FirstSeenBar < 0 && barIdx >= 0 {
		h.FirstSeenBar = barIdx
	}
	if barIdx > h.LastSeenBar {
		h.LastSeenBar = barIdx
	}
	if _, already := warned[k]; !already {
		warned[k] = struct{}{}
		log.Printf("FEATURE-GAP runtime: %s (source=%s, firstBar=%d)", name, source, barIdx)
	}
	mu.Unlock()
	return math.NaN()
}

func RecordStatic(name, source, phase, impact, substitution string, sinks []string) {
	RecordStaticAt(name, source, "", 0, 0, phase, impact, substitution, sinks)
}

func RecordStaticAt(name, source, file string, line, column int, phase, impact, substitution string, sinks []string) {
	mu.Lock()
	k := key{name: name, source: source}
	h, ok := hits[k]
	if !ok {
		h = &Hit{
			Name:         name,
			Source:       source,
			File:         file,
			Line:         line,
			Column:       column,
			Phase:        phase,
			Impact:       impact,
			Substitution: substitution,
			Sinks:        append([]string(nil), sinks...),
			FirstSeenBar: -1,
			LastSeenBar:  -1,
		}
		hits[k] = h
	}
	if impactRank(impact) > impactRank(h.Impact) {
		h.Impact = impact
	}
	if h.File == "" && file != "" {
		h.File = file
		h.Line = line
		h.Column = column
	}
	h.Sinks = mergeStrings(h.Sinks, sinks)
	mu.Unlock()
}

func Snapshot() []Hit {
	mu.Lock()
	defer mu.Unlock()
	out := make([]Hit, 0, len(hits))
	for _, h := range hits {
		out = append(out, *h)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Name != out[j].Name {
			return out[i].Name < out[j].Name
		}
		return out[i].Source < out[j].Source
	})
	return out
}

func HasBacktestCritical(snap []Hit) bool {
	for _, h := range snap {
		if h.Impact == "backtest-critical" {
			return true
		}
	}
	return false
}

func BacktestSinks(snap []Hit) []string {
	var sinks []string
	for _, h := range snap {
		if h.Impact == "backtest-critical" {
			sinks = mergeStrings(sinks, h.Sinks)
		}
	}
	sort.Strings(sinks)
	return sinks
}

func impactRank(impact string) int {
	switch impact {
	case "silent-presentation":
		return 1
	case "observable-non-backtest":
		return 2
	case "backtest-critical":
		return 3
	case "structural":
		return 4
	default:
		return 0
	}
}

func mergeStrings(existing, incoming []string) []string {
	seen := make(map[string]bool, len(existing)+len(incoming))
	out := make([]string, 0, len(existing)+len(incoming))
	for _, value := range existing {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	for _, value := range incoming {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func Reset() {
	mu.Lock()
	hits = map[key]*Hit{}
	warned = map[key]struct{}{}
	mu.Unlock()
}

func Summary() string {
	snap := Snapshot()
	if len(snap) == 0 {
		return ""
	}
	out := ""
	for _, h := range snap {
		out += fmt.Sprintf("  - %s (source=%s, bars=%d..%d, hits=%d)\n",
			h.Name, h.Source, h.FirstSeenBar, h.LastSeenBar, h.HitCount)
	}
	return out
}
