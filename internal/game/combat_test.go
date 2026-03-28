package game

import (
	"math/rand"
	"testing"
)

func TestFireCreatesBulletForPlayerAndEnemy(t *testing.T) {
	g := newPlayingGameForTest()
	g.bullets = nil
	g.player.dir = up
	g.player.turret = up
	g.fire(&g.player, true)
	if len(g.bullets) != 1 || !g.bullets[0].fromPlayer || g.bullets[0].vy >= 0 {
		t.Fatalf("player bullet mismatch")
	}
	e := &tank{x: 100, y: 100, dir: left, turret: left}
	g.fire(e, false)
	if len(g.bullets) != 2 || g.bullets[1].fromPlayer || g.bullets[1].vx >= 0 {
		t.Fatalf("enemy bullet mismatch")
	}
}

func TestFireUsesTurretDirectionForStrafingShot(t *testing.T) {
	g := newPlayingGameForTest()
	g.bullets = nil
	p := &tank{x: 160, y: 180, dir: right, turret: up}
	g.fire(p, true)
	if len(g.bullets) != 1 {
		t.Fatalf("expected one bullet")
	}
	b := g.bullets[0]
	if b.vx != 0 || b.vy >= 0 {
		t.Fatalf("expected upward strafing shot, got vx=%v vy=%v", b.vx, b.vy)
	}
}

func TestUpdateBulletsDamagesFortress(t *testing.T) {
	g := newPlayingGameForTest()
	g.walls = nil
	g.score = 10
	g.bullets = []*bullet{{x: g.fort.box.x + 2, y: g.fort.box.y + 2, vx: 0, vy: 0, fromPlayer: false, alive: true, dmg: 1}}
	hp := g.fort.hp
	g.updateBullets()
	if g.fort.hp != hp-fortHitDamage {
		t.Fatalf("fortress hp should decrease by %d, got %d -> %d", fortHitDamage, hp, g.fort.hp)
	}
	if g.score != 7 {
		t.Fatalf("fortress hit should reduce score by %d, got %d", fortHitLoss, g.score)
	}
}

func TestUpdateBulletsFortressHPClampedAtZero(t *testing.T) {
	g := newPlayingGameForTest()
	g.walls = nil
	g.fort.hp = 1
	g.bullets = []*bullet{{x: g.fort.box.x + 2, y: g.fort.box.y + 2, vx: 0, vy: 0, fromPlayer: false, alive: true, dmg: 1}}
	g.updateBullets()
	if g.fort.hp != 0 {
		t.Fatalf("fortress hp should clamp at zero, got %d", g.fort.hp)
	}
}

func TestUpdateBulletsPlayerShieldAbsorbsDamage(t *testing.T) {
	g := newPlayingGameForTest()
	g.walls = nil
	g.shieldTick = 60
	g.bullets = []*bullet{{x: g.player.x + 2, y: g.player.y + 2, vx: 0, vy: 0, fromPlayer: false, alive: true, dmg: 1}}
	hp := g.player.hp
	g.updateBullets()
	if g.player.hp != hp {
		t.Fatalf("shield should absorb damage")
	}
}

func TestUpdateBulletsPlayerTakesDamageWithoutShield(t *testing.T) {
	g := newPlayingGameForTest()
	g.walls = nil
	g.shieldTick = 0
	g.bullets = []*bullet{{x: g.player.x + 2, y: g.player.y + 2, vx: 0, vy: 0, fromPlayer: false, alive: true, dmg: 1}}
	hp := g.player.hp
	turretHP := g.player.turretHP
	g.updateBullets()
	if g.player.hp != hp-1 {
		t.Fatalf("player should take damage")
	}
	if g.player.turretHP != turretHP-1 {
		t.Fatalf("player turret should take damage")
	}
}

func TestUpdateBulletsKillsEnemyAndScores(t *testing.T) {
	g := newPlayingGameForTest()
	g.walls = nil
	g.enemies = []*tank{{x: 200, y: 200, hp: 1}}
	g.bullets = []*bullet{{x: 202, y: 202, vx: 0, vy: 0, fromPlayer: true, alive: true, dmg: 1}}
	rand.Seed(999)
	g.updateBullets()
	if len(g.enemies) != 0 || g.score < 150 {
		t.Fatalf("enemy should be removed and score increased")
	}
}

func TestUpdateBulletsGuardWallHitLosesScore(t *testing.T) {
	g := newPlayingGameForTest()
	g.score = 9
	g.walls = []*wall{{box: rect{x: 120, y: 120, w: 10, h: 10}, hp: 1, maxHP: 1, destructive: true, guard: true}}
	g.bullets = []*bullet{{x: 122, y: 122, vx: 0, vy: 0, fromPlayer: false, alive: true, dmg: 1}}
	g.updateBullets()
	if g.score != 7 {
		t.Fatalf("guard wall hit should reduce score by %d, got %d", guardHitLoss, g.score)
	}
	if g.walls[0].hp != 0 {
		t.Fatalf("guard wall chunk should be removed in one hit")
	}
}

func TestUpdateBulletsScoreLossClampedAtZero(t *testing.T) {
	g := newPlayingGameForTest()
	g.score = 1
	g.walls = nil
	g.bullets = []*bullet{{x: g.fort.box.x + 1, y: g.fort.box.y + 1, vx: 0, vy: 0, fromPlayer: false, alive: true, dmg: 1}}
	g.updateBullets()
	if g.score != 0 {
		t.Fatalf("score should be clamped at zero, got %d", g.score)
	}
}

func TestGuardWallNotDestroyedInOneHit(t *testing.T) {
	g := newPlayingGameForTest()
	g.walls = []*wall{
		{box: rect{x: 120, y: 120, w: 10, h: 10}, hp: 1, maxHP: 1, destructive: true, guard: true},
		{box: rect{x: 130, y: 120, w: 10, h: 10}, hp: 1, maxHP: 1, destructive: true, guard: true},
	}
	g.bullets = []*bullet{{x: 122, y: 122, vx: 0, vy: 0, fromPlayer: false, alive: true, dmg: 1}}
	g.updateBullets()
	g.cleanupWalls()
	if len(g.walls) != 1 {
		t.Fatalf("guard wall should lose only one square chunk per hit")
	}
}

func TestFortGuardWallNeedsTwoHitsToDestroy(t *testing.T) {
	g := newPlayingGameForTest()
	var target *wall
	for _, w := range g.walls {
		if w.guard {
			target = w
			break
		}
	}
	if target == nil {
		t.Fatalf("expected at least one fortress guard wall chunk")
	}
	px := target.box.x + target.box.w/2
	py := target.box.y + target.box.h/2
	g.bullets = []*bullet{{x: px, y: py, vx: 0, vy: 0, fromPlayer: false, alive: true, dmg: 1}}
	g.updateBullets()
	if target.hp != 1 {
		t.Fatalf("first hit should reduce guard hp to 1, got %d", target.hp)
	}
	g.cleanupWalls()
	if target.hp <= 0 {
		t.Fatalf("guard chunk should still exist after first hit")
	}

	g.bullets = []*bullet{{x: px, y: py, vx: 0, vy: 0, fromPlayer: false, alive: true, dmg: 1}}
	g.updateBullets()
	g.cleanupWalls()
	if target.hp > 0 {
		t.Fatalf("second hit should destroy guard chunk, got hp=%d", target.hp)
	}
}

func TestObstacleChunkOneHitRemovesOnlyOnePiece(t *testing.T) {
	g := newPlayingGameForTest()
	g.walls = []*wall{
		{box: rect{x: 100, y: 100, w: 16, h: 16}, hp: 1, maxHP: 1, destructive: true, guard: false},
		{box: rect{x: 116, y: 100, w: 16, h: 16}, hp: 1, maxHP: 1, destructive: true, guard: false},
	}
	g.bullets = []*bullet{{x: 102, y: 102, vx: 0, vy: 0, fromPlayer: true, alive: true, dmg: 1}}

	g.updateBullets()
	g.cleanupWalls()

	if len(g.walls) != 1 {
		t.Fatalf("expected only one obstacle chunk removed, left=%d", len(g.walls))
	}
	if g.walls[0].box.x != 116 {
		t.Fatalf("wrong chunk remained after one hit")
	}
}

func TestEnemyFireIsAxisAligned(t *testing.T) {
	g := newPlayingGameForTest()
	g.bullets = nil
	e := &tank{x: 120, y: 120}
	for _, d := range []direction{up, down, left, right} {
		e.dir = up
		e.turret = d
		g.fire(e, false)
		b := g.bullets[len(g.bullets)-1]
		// Enemy bullets must stay axis-aligned (no diagonal components).
		if b.vx != 0 && b.vy != 0 {
			t.Fatalf("enemy bullet should not be diagonal, dir=%v vx=%v vy=%v", d, b.vx, b.vy)
		}
	}
}

func TestUpdateBulletsOppositeBulletsCancelEachOther(t *testing.T) {
	g := newPlayingGameForTest()
	g.walls = nil
	g.enemies = nil
	g.player.x = 40
	g.player.y = 40
	playerHP := g.player.hp
	fortHP := g.fort.hp

	g.bullets = []*bullet{
		{x: 300, y: 300, vx: 0, vy: 0, fromPlayer: true, alive: true, dmg: 1},
		{x: 300, y: 300, vx: 0, vy: 0, fromPlayer: false, alive: true, dmg: 1},
	}

	g.updateBullets()

	if len(g.bullets) != 0 {
		t.Fatalf("opposite bullets should cancel out, remaining=%d", len(g.bullets))
	}
	if g.player.hp != playerHP {
		t.Fatalf("player hp should stay unchanged after bullet clash")
	}
	if g.fort.hp != fortHP {
		t.Fatalf("fortress hp should stay unchanged after bullet clash")
	}
}
