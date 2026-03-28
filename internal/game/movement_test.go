package game

import (
	"math/rand"
	"testing"
)

func TestTryMoveTankDiagonalSlide(t *testing.T) {
	g := newGame()
	g.startMatch()
	g.enemies = nil
	g.player.x = 100
	g.player.y = 100
	g.walls = append(g.walls, &wall{box: rect{x: 134, y: 134, w: 20, h: 20}, hp: 99, maxHP: 99, destructive: false})

	ok := g.tryMoveTank(&g.player, 3, 3)
	if !ok {
		t.Fatalf("expected slide movement to succeed")
	}
	if g.player.y != 100 {
		t.Fatalf("expected Y to stay due to diagonal block, got y=%v", g.player.y)
	}
	if g.player.x <= 100 {
		t.Fatalf("expected X to advance, got x=%v", g.player.x)
	}
}

func TestCanOccupyRejectsEnemyOverlap(t *testing.T) {
	g := newGame()
	g.startMatch()
	e := g.enemies[0]
	next := tankRect(*e)
	if g.canOccupy(next, &g.player) {
		t.Fatalf("player should not occupy enemy position")
	}
}

func TestCanOccupyRejectsOutOfMap(t *testing.T) {
	g := newPlayingGameForTest()
	if g.canOccupy(rect{-1, 20, tankSize, tankSize}, &g.player) {
		t.Fatalf("negative x should be rejected")
	}
	if g.canOccupy(rect{20, -1, tankSize, tankSize}, &g.player) {
		t.Fatalf("negative y should be rejected")
	}
}

func TestCanOccupyRejectsFortress(t *testing.T) {
	g := newPlayingGameForTest()
	if g.canOccupy(g.fort.box, &g.player) {
		t.Fatalf("fortress area must be blocked")
	}
}

func TestCanOccupyRejectsWall(t *testing.T) {
	g := newPlayingGameForTest()
	if g.canOccupy(g.walls[0].box, &g.player) {
		t.Fatalf("wall area must be blocked")
	}
}

func TestPlaceEnemyFailsWhenFullyBlocked(t *testing.T) {
	g := newPlayingGameForTest()
	g.walls = []*wall{{box: rect{0, 0, screenW, screenH}, hp: 99, maxHP: 99, destructive: false}}
	e := &tank{}
	if g.placeEnemy(e, 100, 100) {
		t.Fatalf("placeEnemy should fail in fully blocked map")
	}
}

func TestEnemyCrowdedAtTrueAndFalse(t *testing.T) {
	g := newPlayingGameForTest()
	g.enemies = []*tank{{x: 100, y: 100}, {x: 120, y: 100}}
	if !g.enemyCrowdedAt(g.enemies[0], 117, 117, 60) {
		t.Fatalf("expected crowded")
	}
	if g.enemyCrowdedAt(g.enemies[0], 400, 400, 60) {
		t.Fatalf("expected not crowded")
	}
}

func TestNextStepByDirection(t *testing.T) {
	x, y := nextStep(10, 10, up, 2)
	if x != 10 || y != 8 {
		t.Fatalf("up step mismatch")
	}
	x, y = nextStep(10, 10, right, 3)
	if x != 13 || y != 10 {
		t.Fatalf("right step mismatch")
	}
}

func TestDirVectorAndOpposite(t *testing.T) {
	x, y := dirVector(left)
	if x != -1 || y != 0 {
		t.Fatalf("dirVector left mismatch")
	}
	if oppositeDir(up) != down || oppositeDir(right) != left {
		t.Fatalf("oppositeDir mismatch")
	}
}

func TestClampF(t *testing.T) {
	if clampF(2, 3, 5) != 3 || clampF(8, 3, 5) != 5 || clampF(4, 3, 5) != 4 {
		t.Fatalf("clampF mismatch")
	}
}

func TestTryMoveTankZeroVectorFalse(t *testing.T) {
	g := newPlayingGameForTest()
	if g.tryMoveTank(&g.player, 0, 0) {
		t.Fatalf("zero movement should return false")
	}
}

func TestTryUnstuckFindsAlternativeDirection(t *testing.T) {
	g := newPlayingGameForTest()
	e := &tank{x: 100, y: 100, dir: up, speed: 2}
	g.enemies = []*tank{e}
	g.walls = []*wall{{box: rect{100, 95, 34, 4}, hp: 99, maxHP: 99, destructive: false}}
	if !g.tryUnstuck(e) {
		t.Fatalf("tryUnstuck should find alternative")
	}
}

func TestFlowDirForEnemyReturnsValidDirection(t *testing.T) {
	g := newPlayingGameForTest()
	d := g.flowDirForEnemy(g.enemies[0])
	if d != up && d != down && d != left && d != right {
		t.Fatalf("invalid direction")
	}
}

func TestApplyEnemyTurnNeedsMoreVotesForOpposite(t *testing.T) {
	g := newPlayingGameForTest()
	e := &tank{dir: left}
	g.applyEnemyTurn(e, right)
	g.applyEnemyTurn(e, right)
	g.applyEnemyTurn(e, right)
	g.applyEnemyTurn(e, right)
	g.applyEnemyTurn(e, right)
	g.applyEnemyTurn(e, right)
	if e.dir != left {
		t.Fatalf("opposite turn should not commit before required votes")
	}
	g.applyEnemyTurn(e, right)
	if e.dir != right {
		t.Fatalf("opposite turn should commit once vote threshold is reached")
	}
}

func TestApplyEnemyTurnAllowsQuickTurnWhenStuck(t *testing.T) {
	g := newPlayingGameForTest()
	e := &tank{dir: up, stuck: 2}
	g.applyEnemyTurn(e, down)
	if e.dir != down {
		t.Fatalf("stuck enemy should be allowed to reverse quickly")
	}
}

func TestApplyEnemyTurnRespectsTurnLock(t *testing.T) {
	g := newPlayingGameForTest()
	e := &tank{dir: up, turnLock: 3}
	g.applyEnemyTurn(e, left)
	g.applyEnemyTurn(e, left)
	g.applyEnemyTurn(e, left)
	if e.dir != up {
		t.Fatalf("enemy should not turn while turn lock is active")
	}
}

func TestFlowDirForEnemyKeepsCurrentDirWhenGainIsSmall(t *testing.T) {
	g := newPlayingGameForTest()
	g.walls = nil
	e := &tank{x: 300, y: 300, dir: up, speed: 2, role: roleAssault}
	g.enemies = []*tank{e}

	// Place fortress almost directly above to keep "up" competitive.
	g.fort.box = rect{x: 300, y: 220, w: 64, h: 30}
	// Place player to the left to add a slight rightward force, but not dominant.
	g.player.x = 250
	g.player.y = 300

	d := g.flowDirForEnemy(e)
	if d != up {
		t.Fatalf("expected hysteresis to keep current direction, got %v", d)
	}
}

func TestEnemyFireChanceIncreasesWhenAligned(t *testing.T) {
	base := enemyFireChance(false, false, 0, 0.5)
	alignedBase := enemyFireChance(true, false, 0, 0.5)
	alignedPlayer := enemyFireChance(false, true, 0, 0.5)
	if alignedBase <= base {
		t.Fatalf("aligned base should increase fire chance")
	}
	if alignedPlayer <= base {
		t.Fatalf("aligned player should increase fire chance")
	}
}

func TestEnemyFireChanceClamped(t *testing.T) {
	low := enemyFireChance(false, false, -100, 0.0)
	high := enemyFireChance(true, true, 200, 1.0)
	if low != 1 {
		t.Fatalf("fire chance lower clamp mismatch, got %d", low)
	}
	if high != 70 {
		t.Fatalf("fire chance upper clamp mismatch, got %d", high)
	}
}

func TestEnemyFireCooldownRange(t *testing.T) {
	rand.Seed(7)
	for i := 0; i < 50; i++ {
		v := enemyFireCooldown(false, false, 0.5)
		if v < enemyFireCooldownBaseMin || v >= enemyFireCooldownBaseMin+enemyFireCooldownBaseVar {
			t.Fatalf("cooldown out of expected range: %d", v)
		}
	}
}

func TestPlayerFireCooldownValues(t *testing.T) {
	if playerFireCooldown(false) != playerFireCooldownFrames {
		t.Fatalf("normal fire cooldown mismatch")
	}
	if playerFireCooldown(true) != playerRapidFireCooldownFrames {
		t.Fatalf("rapid fire cooldown mismatch")
	}
	if playerRapidFireCooldownFrames >= playerFireCooldownFrames {
		t.Fatalf("rapid cooldown should be shorter than normal cooldown")
	}
}

func TestNextEnemyPlanDirReturnsOccupiableDirection(t *testing.T) {
	rand.Seed(11)
	g := newPlayingGameForTest()
	g.walls = nil
	e := &tank{x: 200, y: 200, dir: up, speed: 2}
	g.enemies = []*tank{e}
	d := g.nextEnemyPlanDir(e)
	sx, sy := nextStep(e.x, e.y, d, e.speed)
	if !g.canOccupy(rect{x: sx, y: sy, w: tankSize, h: tankSize}, e) {
		t.Fatalf("planned direction should be occupiable")
	}
}

func TestInitEnemyTraitsAddsPerTankRandomFactors(t *testing.T) {
	rand.Seed(23)
	e1 := &tank{}
	e2 := &tank{}
	initEnemyTraits(e1)
	initEnemyTraits(e2)
	if e1.aiRand == e2.aiRand && e1.replan == e2.replan && e1.fireBias == e2.fireBias && e1.aggro == e2.aggro {
		t.Fatalf("expected per-tank randomized traits, but both are identical")
	}
}

func TestEnemyTargetPointConvergesToFortressOverTime(t *testing.T) {
	g := newPlayingGameForTest()
	e := &tank{role: roleLeftFlank, aiRand: 0.5, aggro: 1}
	cx, cy := 200.0, 100.0

	e.age = 0
	xEarly, _ := g.enemyTargetPoint(e, cx, cy)
	e.age = 2000
	xLate, _ := g.enemyTargetPoint(e, cx, cy)
	baseX := g.fort.box.x + g.fort.box.w/2
	if mathAbs(baseX-xLate) >= mathAbs(baseX-xEarly) {
		t.Fatalf("expected target point to move closer to fortress over time")
	}
}

func TestEnemyFortressHitProbabilityIncreasesWithSilence(t *testing.T) {
	g := newPlayingGameForTest()
	e := &tank{aggro: 1, age: 300}
	g.playerSilentFrames = 0
	low := g.enemyFortressHitProbability(e, false)
	g.playerSilentFrames = 1800
	high := g.enemyFortressHitProbability(e, false)
	if high <= low {
		t.Fatalf("expected fortress hit probability to increase with player silence: low=%d high=%d", low, high)
	}
}

func TestEnemyTargetPointConvergesWithSilenceEvenWithoutKills(t *testing.T) {
	g := newPlayingGameForTest()
	e := &tank{role: roleRightFlank, aiRand: 0.7, aggro: 1, age: 0}
	cx, cy := 300.0, 140.0
	baseX := g.fort.box.x + g.fort.box.w/2

	g.playerSilentFrames = 0
	xQuietStart, _ := g.enemyTargetPoint(e, cx, cy)
	g.playerSilentFrames = 1800
	xQuietLong, _ := g.enemyTargetPoint(e, cx, cy)
	if mathAbs(baseX-xQuietLong) >= mathAbs(baseX-xQuietStart) {
		t.Fatalf("expected silence pressure to push target point toward fortress")
	}
}

func TestEnemyTargetPointKeepsSpreadUnderHighPressure(t *testing.T) {
	g := newPlayingGameForTest()
	g.playerSilentFrames = 99999
	e1 := &tank{role: roleLeftFlank, aiRand: 0.1, aggro: 1, age: 3000}
	e2 := &tank{role: roleLeftFlank, aiRand: 0.9, aggro: 1, age: 3000}
	x1, _ := g.enemyTargetPoint(e1, 200, 120)
	x2, _ := g.enemyTargetPoint(e2, 200, 120)
	if mathAbs(x1-x2) < 12 {
		t.Fatalf("high pressure should still keep target spread, got |dx|=%v", mathAbs(x1-x2))
	}
}

func TestCrowdPenaltyHigherWhenCloser(t *testing.T) {
	g := newPlayingGameForTest()
	self := &tank{x: 200, y: 200, aiRand: 0.5}
	other := &tank{x: 220, y: 200}
	g.enemies = []*tank{self, other}
	near := g.crowdPenaltyAt(self, 220, 220)
	far := g.crowdPenaltyAt(self, 420, 420)
	if near <= far {
		t.Fatalf("expected crowd penalty to be higher for near positions")
	}
}

func TestOnPlayerFiredResetsSilenceCounter(t *testing.T) {
	g := newPlayingGameForTest()
	g.playerSilentFrames = 300
	g.onPlayerFired()
	if g.playerSilentFrames != 0 {
		t.Fatalf("player silence counter should reset on fire")
	}
}

func TestOnPlayerDirTapNeedsDoubleTapToTurn(t *testing.T) {
	g := newPlayingGameForTest()
	g.player.dir = up
	g.player.turret = up

	g.frame = 10
	if g.onPlayerDirTap(left) {
		t.Fatalf("first tap should not turn")
	}
	if g.player.dir != up || g.player.turret != up {
		t.Fatalf("orientation should stay unchanged after first tap")
	}

	g.frame = 19
	if !g.onPlayerDirTap(left) {
		t.Fatalf("second tap within window should turn")
	}
	if g.player.dir != left || g.player.turret != left {
		t.Fatalf("double-tap should sync body and turret direction")
	}
}

func TestOnPlayerDirTapOutsideWindowDoesNotTurn(t *testing.T) {
	g := newPlayingGameForTest()
	g.player.dir = up
	g.player.turret = up

	g.frame = 20
	g.onPlayerDirTap(right)
	g.frame = 20 + playerTurnDoubleTapFrames + 1
	if g.onPlayerDirTap(right) {
		t.Fatalf("tap outside double-tap window should not turn")
	}
	if g.player.dir != up || g.player.turret != up {
		t.Fatalf("direction should remain unchanged when window exceeded")
	}
}

func TestOnPlayerDirTapSetsMoveLock(t *testing.T) {
	g := newPlayingGameForTest()
	g.frame = 40
	g.onPlayerDirTap(up)
	g.frame = 50
	if !g.onPlayerDirTap(up) {
		t.Fatalf("expected second tap to trigger turn")
	}
	if g.playerMoveLockUntil != g.frame+playerTurnMoveLockFrames {
		t.Fatalf("move lock mismatch: got %d", g.playerMoveLockUntil)
	}
}

func TestCanMoveOnHeldRespectsGraceAndMoveLock(t *testing.T) {
	g := newPlayingGameForTest()
	g.playerPressStart[up] = 10

	g.frame = 13
	if g.canMoveOnHeld(up, true) {
		t.Fatalf("should not move before tap grace")
	}

	g.frame = 20
	g.playerMoveLockUntil = 25
	if g.canMoveOnHeld(up, true) {
		t.Fatalf("should not move while turn move lock is active")
	}

	g.frame = 26
	if !g.canMoveOnHeld(up, true) {
		t.Fatalf("should move after grace and lock window")
	}
}

func TestCanMoveOnHeldBootstrapsWhenPressedWithoutJustPressed(t *testing.T) {
	g := newPlayingGameForTest()
	g.resetPlayerTapFrames()
	g.frame = 100
	if g.canMoveOnHeld(right, true) {
		t.Fatalf("first observed pressed frame should be treated as grace window")
	}
	g.frame = 100 + playerTapGraceFrames
	if !g.canMoveOnHeld(right, true) {
		t.Fatalf("held input should move after grace even without just-pressed signal")
	}
}

func TestApplyPlayerTurnTapsDoubleTapTurns(t *testing.T) {
	g := newPlayingGameForTest()
	g.player.dir = up
	g.player.turret = up

	g.frame = 30
	g.applyPlayerTurnTaps(false, false, false, true)
	if g.player.dir != up || g.player.turret != up {
		t.Fatalf("single tap should not turn")
	}

	g.frame = 40
	g.applyPlayerTurnTaps(false, false, false, true)
	if g.player.dir != right || g.player.turret != right {
		t.Fatalf("double tap should turn to right")
	}
}

func TestApplyPlayerTurnTapsOutsideWindowNoTurn(t *testing.T) {
	g := newPlayingGameForTest()
	g.player.dir = left
	g.player.turret = left

	g.frame = 10
	g.applyPlayerTurnTaps(true, false, false, false)
	g.frame = 10 + playerTurnDoubleTapFrames + 1
	g.applyPlayerTurnTaps(true, false, false, false)
	if g.player.dir != left || g.player.turret != left {
		t.Fatalf("tap outside window should not change facing")
	}
}

func TestSingleEnemySilenceEventuallyDamagesFortress(t *testing.T) {
	rand.Seed(5)
	g := newPlayingGameForTest()
	g.walls = nil
	g.bullets = nil
	g.playerSilentFrames = 99999
	g.player.x = 40
	g.player.y = 40
	e := &tank{
		x:      screenW/2 - tankSize/2,
		y:      70,
		dir:    down,
		speed:  0,
		hp:     2,
		aiRand: 0.5,
		replan: 8,
		aggro:  1,
	}
	g.enemies = []*tank{e}
	start := g.fort.hp
	for i := 0; i < 2000 && g.fort.hp == start; i++ {
		g.updateEnemies()
		g.updateBullets()
	}
	if g.fort.hp >= start {
		t.Fatalf("single enemy should eventually damage fortress when player stays silent")
	}
}

func mathAbs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
