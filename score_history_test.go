package tankbattle

import (
	"testing"
	"time"
)

func TestCurrentRankAndBestScore(t *testing.T) {
	g := newPlayingGameForTest()
	g.scoreHistory = []scoreEntry{
		{Score: 900, At: time.Now().UTC().Format(time.RFC3339)},
		{Score: 700, At: time.Now().UTC().Add(-time.Minute).Format(time.RFC3339)},
		{Score: 200, At: time.Now().UTC().Add(-2 * time.Minute).Format(time.RFC3339)},
	}
	g.score = 650
	if g.bestScore() != 900 {
		t.Fatalf("best score mismatch, got %d", g.bestScore())
	}
	if g.currentRank() != 3 {
		t.Fatalf("current rank mismatch, got %d", g.currentRank())
	}
}

func TestVisibleRankEntriesUsesScrollWindow(t *testing.T) {
	g := newPlayingGameForTest()
	g.scoreHistory = nil
	for i := 0; i < 10; i++ {
		g.scoreHistory = append(g.scoreHistory, scoreEntry{
			Score: 1000 - i,
			At:    time.Now().UTC().Add(time.Duration(-i) * time.Minute).Format(time.RFC3339),
		})
	}
	g.rankScroll = 3
	entries, start := g.visibleRankEntries()
	if start != 3 {
		t.Fatalf("start mismatch, got %d", start)
	}
	if len(entries) != hudRankingRows {
		t.Fatalf("visible rows mismatch, got %d", len(entries))
	}
	if entries[0].Score != 997 {
		t.Fatalf("first visible score mismatch, got %d", entries[0].Score)
	}
}

func TestAppendCurrentScoreHistoryOnlyOnce(t *testing.T) {
	g := newPlayingGameForTest()
	g.scoreHistory = nil
	g.score = 321
	g.matchLogged = false
	g.appendCurrentScoreHistory()
	g.appendCurrentScoreHistory()
	if len(g.scoreHistory) != 1 {
		t.Fatalf("score history should append once, got %d", len(g.scoreHistory))
	}
	if g.scoreHistory[0].Score != 321 {
		t.Fatalf("saved score mismatch, got %d", g.scoreHistory[0].Score)
	}
}
