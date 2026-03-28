package ui

import "testing"

func TestFillCompatGameReusesBackingSlices(t *testing.T) {
	snapshot := Snapshot{
		State: "playing",
		Enemies: []Tank{
			{X: 10, Y: 12, HP: 2, MaxHP: 2},
			{X: 30, Y: 40, HP: 2, MaxHP: 2},
		},
		Walls: []Wall{
			{Box: Rect{X: 10, Y: 10, W: 30, H: 30}, HP: 1, MaxHP: 1, Destructive: true},
		},
		ScoreHistory: []ScoreEntry{
			{Score: 50, DurationSec: 10},
		},
	}

	g := &game{}
	fillCompatGame(g, snapshot)
	if len(g.enemies) != 2 || len(g.walls) != 1 || len(g.scoreHistory) != 1 {
		t.Fatalf("unexpected compat sizes after first fill")
	}
	enemyPtr := &g.enemies[0]
	wallPtr := &g.walls[0]

	fillCompatGame(g, snapshot)
	if &g.enemies[0] != enemyPtr {
		t.Fatalf("enemy compat objects should be reused")
	}
	if &g.walls[0] != wallPtr {
		t.Fatalf("wall compat objects should be reused")
	}
}
