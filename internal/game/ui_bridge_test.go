package game

import "testing"

func TestUISnapshotReusesBackingSlices(t *testing.T) {
	g := newPlayingGameForTest()
	g.scoreHistory = append(g.scoreHistory,
		scoreEntry{Score: 100, DurationSec: 11},
		scoreEntry{Score: 220, DurationSec: 22},
	)

	first := g.uiSnapshot()
	if len(first.Enemies) == 0 || len(first.Walls) == 0 {
		t.Fatalf("expected populated snapshot slices")
	}

	enemyPtr := &first.Enemies[0]
	wallPtr := &first.Walls[0]
	historyPtr := &first.ScoreHistory[0]

	second := g.uiSnapshot()
	if &second.Enemies[0] != enemyPtr {
		t.Fatalf("enemy snapshot backing array should be reused")
	}
	if &second.Walls[0] != wallPtr {
		t.Fatalf("wall snapshot backing array should be reused")
	}
	if &second.ScoreHistory[0] != historyPtr {
		t.Fatalf("score history snapshot backing array should be reused")
	}
}
