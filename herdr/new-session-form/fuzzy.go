package main

import (
	"sort"
	"strings"
	"unicode"
)

// fuzzyMatch scores pattern as a subsequence of target, fzf-style: consecutive
// runs and word starts score up, gaps score down, so "prn" prefers "PRN
// initials…" over "Pong response". Case-insensitive with a nudge for exact
// case. Every start position for the first pattern rune is tried and the best
// greedy alignment from it wins. Returns ok=false when pattern isn't a
// subsequence at all.
func fuzzyMatch(pattern, target string) (score int, matched []int, ok bool) {
	p := []rune(pattern)
	t := []rune(target)
	if len(p) == 0 {
		return 0, nil, true
	}
	tl := make([]rune, len(t))
	for i, r := range t {
		tl[i] = unicode.ToLower(r)
	}
	pl := make([]rune, len(p))
	for i, r := range p {
		pl[i] = unicode.ToLower(r)
	}

	const (
		scoreMatch       = 16
		bonusConsecutive = 8
		bonusBoundary    = 8
		bonusExactCase   = 1
		penaltyGapStart  = 3
		penaltyGapExtend = 1
	)

	best := -1 << 30
	var bestIdx []int
	for start := 0; start < len(t); start++ {
		if tl[start] != pl[0] {
			continue
		}
		idx := make([]int, 0, len(p))
		idx = append(idx, start)
		pos := start
		for pi := 1; pi < len(p); pi++ {
			next := -1
			for j := pos + 1; j < len(t); j++ {
				if tl[j] == pl[pi] {
					next = j
					break
				}
			}
			if next == -1 {
				idx = nil
				break
			}
			idx = append(idx, next)
			pos = next
		}
		if idx == nil {
			continue
		}

		s := 0
		prev := -2
		for k, i := range idx {
			s += scoreMatch
			if i == prev+1 {
				s += bonusConsecutive
			} else if k > 0 {
				gap := i - prev - 1
				s -= penaltyGapStart + penaltyGapExtend*(gap-1)
			}
			if isBoundary(t, i) {
				s += bonusBoundary
			}
			if t[i] == p[k] {
				s += bonusExactCase
			}
			prev = i
		}
		if s > best {
			best, bestIdx = s, idx
		}
	}
	if bestIdx == nil {
		return 0, nil, false
	}
	return best, bestIdx, true
}

// isBoundary: start of string, after a non-alphanumeric, or an uppercase rune
// following a lowercase one (camelCase).
func isBoundary(t []rune, i int) bool {
	if i == 0 {
		return true
	}
	prev := t[i-1]
	if !unicode.IsLetter(prev) && !unicode.IsDigit(prev) {
		return true
	}
	return unicode.IsUpper(t[i]) && unicode.IsLower(prev)
}

type fuzzyHit struct {
	index   int // position in the input slice
	score   int
	matched []int
}

// fuzzyRank returns the targets that match, best score first; ties keep the
// input order (which callers pass as most-recent-first). Whitespace splits
// the pattern into terms that must each match somewhere (fzf's extended
// mode), so "emar rev" means emar AND rev, not a literal space.
func fuzzyRank(pattern string, targets []string) []fuzzyHit {
	terms := strings.Fields(pattern)
	var hits []fuzzyHit
	for i, target := range targets {
		total := 0
		var all []int
		ok := true
		for _, term := range terms {
			score, matched, found := fuzzyMatch(term, target)
			if !found {
				ok = false
				break
			}
			total += score
			all = append(all, matched...)
		}
		if ok {
			hits = append(hits, fuzzyHit{index: i, score: total, matched: all})
		}
	}
	sort.SliceStable(hits, func(a, b int) bool { return hits[a].score > hits[b].score })
	return hits
}
