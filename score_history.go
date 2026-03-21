package tankbattle

import "time"

func (g *game) bestScore() int {
	best := 0
	for _, e := range g.scoreHistory {
		if e.Score > best {
			best = e.Score
		}
	}
	if g.score > best {
		best = g.score
	}
	return best
}

func (g *game) currentRank() int {
	rank := 1
	for _, e := range g.scoreHistory {
		if e.Score > g.score {
			rank++
		}
	}
	return rank
}

func (g *game) maxRankScroll() int {
	if len(g.scoreHistory) <= hudRankingRows {
		return 0
	}
	return len(g.scoreHistory) - hudRankingRows
}

func (g *game) clampRankScroll() {
	if g.rankScroll < 0 {
		g.rankScroll = 0
	}
	maxScroll := g.maxRankScroll()
	if g.rankScroll > maxScroll {
		g.rankScroll = maxScroll
	}
}

func (g *game) visibleRankEntries() ([]scoreEntry, int) {
	if len(g.scoreHistory) == 0 {
		return nil, 0
	}
	g.clampRankScroll()
	start := g.rankScroll
	end := start + hudRankingRows
	if end > len(g.scoreHistory) {
		end = len(g.scoreHistory)
	}
	return g.scoreHistory[start:end], start
}

func (g *game) appendCurrentScoreHistory() {
	if g == nil || g.matchLogged {
		return
	}
	g.matchLogged = true
	g.scoreHistory = append(g.scoreHistory, scoreEntry{
		Score: maxInt(g.score, 0),
		At:    time.Now().UTC().Format(time.RFC3339),
	})
	g.scoreHistory = sanitizeScoreHistory(g.scoreHistory)
	g.clampRankScroll()
	g.saveUserSettings()
}
