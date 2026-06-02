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
	Name         string
	Source       string
	FirstSeenBar int
	LastSeenBar  int
	HitCount     int
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
	mu.Lock()
	k := key{name: name, source: source}
	h, ok := hits[k]
	if !ok {
		h = &Hit{Name: name, Source: source, FirstSeenBar: barIdx, LastSeenBar: barIdx}
		hits[k] = h
	}
	h.HitCount++
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
