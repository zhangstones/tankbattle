package game

import (
	"math"
	"testing"
)

func TestCleanupWallsRemovesDestroyedDestructive(t *testing.T) {
	g := newPlayingGameForTest()
	g.walls = []*wall{
		{box: rect{0, 0, 10, 10}, hp: 0, maxHP: 2, destructive: true},
		{box: rect{20, 0, 10, 10}, hp: 0, maxHP: 2, destructive: false},
	}
	g.cleanupWalls()
	if len(g.walls) != 1 || g.walls[0].destructive {
		t.Fatalf("cleanupWalls should keep only non-destructive wall")
	}
}

func TestUpdateExplosionsDecay(t *testing.T) {
	g := newPlayingGameForTest()
	g.explosions = []*explosion{{x: 1, y: 1, radius: 3, life: 1, max: 2}}
	g.updateExplosions()
	if len(g.explosions) != 0 {
		t.Fatalf("explosion with life 1 should disappear")
	}
}

func TestPhysicsHelpers(t *testing.T) {
	if !overlap(rect{0, 0, 10, 10}, rect{9, 9, 10, 10}) {
		t.Fatalf("overlap expected true")
	}
	if overlap(rect{0, 0, 10, 10}, rect{20, 20, 5, 5}) {
		t.Fatalf("overlap expected false")
	}
	tr := tankRect(tank{x: 3, y: 4})
	if math.Abs(tr.x-3) > 0.001 || math.Abs(tr.y-4) > 0.001 {
		t.Fatalf("tankRect mismatch")
	}
}
