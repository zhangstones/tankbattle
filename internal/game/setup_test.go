package game

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
)

func TestStartMatchResetsCoreState(t *testing.T) {
	g := newGame()
	g.score = 999
	g.wave = 3
	g.win = true
	g.paused = true
	g.state = stateEnded
	g.startMatch()
	if g.state != statePlaying || g.wave != 1 || g.score != 0 || g.win || g.paused {
		t.Fatalf("startMatch should reset match state")
	}
	if g.fort.hp != fortressMaxHP || g.fort.maxHP != fortressMaxHP {
		t.Fatalf("fortress hp should reset to %d/%d", fortressMaxHP, fortressMaxHP)
	}
	if g.player.hp != playerHullMaxHP || g.player.maxHP != playerHullMaxHP {
		t.Fatalf("player hull hp should reset to %d/%d", playerHullMaxHP, playerHullMaxHP)
	}
	if g.player.turretHP != playerTurretMaxHP || g.player.turretMaxHP != playerTurretMaxHP {
		t.Fatalf("player turret hp should reset to %d/%d", playerTurretMaxHP, playerTurretMaxHP)
	}
	if g.matchIntroTick == 0 || g.matchIntroTick != g.matchIntroMax {
		t.Fatalf("startMatch should begin intro transition, tick=%d max=%d", g.matchIntroTick, g.matchIntroMax)
	}
}

func TestStartMatchResamplesBackgroundSeed(t *testing.T) {
	seed := int64(321)
	g := newGameWithOptions(newGameOptions{
		loadUserSettings: false,
		persistUserData:  false,
		randomSeed:       &seed,
	})
	initial := g.backgroundSeed
	g.startMatch()
	first := g.backgroundSeed
	g.startMatch()
	second := g.backgroundSeed
	if initial == 0 || first == 0 || second == 0 {
		t.Fatalf("background seed should always be initialized")
	}
	if first == initial {
		t.Fatalf("startMatch should resample background seed from menu seed")
	}
	if second == first {
		t.Fatalf("restarting should resample background seed, got repeated %d", second)
	}
}

func TestMaxWaveByDifficulty(t *testing.T) {
	g := newGame()
	g.difficulty = diffEasy
	if g.maxWaveByDifficulty() != 3 {
		t.Fatalf("easy maxWave mismatch")
	}
	g.difficulty = diffNormal
	if g.maxWaveByDifficulty() != 4 {
		t.Fatalf("normal maxWave mismatch")
	}
	g.difficulty = diffHard
	if g.maxWaveByDifficulty() != 5 {
		t.Fatalf("hard maxWave mismatch")
	}
}

func TestClampedTotalWaves(t *testing.T) {
	g := newGame()
	g.totalWaves = -3
	if g.clampedTotalWaves() != matchWaveMin {
		t.Fatalf("total waves should clamp to min %d", matchWaveMin)
	}
	g.totalWaves = 999
	if g.clampedTotalWaves() != matchWaveMax {
		t.Fatalf("total waves should clamp to max %d", matchWaveMax)
	}
}

func TestStartMatchUsesConfiguredTotalWaves(t *testing.T) {
	g := newGame()
	g.difficulty = diffHard
	g.totalWaves = 1
	g.startMatch()
	if g.maxWave != 1 {
		t.Fatalf("startMatch should use configured total waves, got %d", g.maxWave)
	}
}

func TestEnemyBaseByDifficulty(t *testing.T) {
	g := newGame()
	g.difficulty = diffEasy
	if g.enemyBaseByDifficulty() != 2 {
		t.Fatalf("easy enemy base mismatch")
	}
	g.difficulty = diffNormal
	if g.enemyBaseByDifficulty() != 3 {
		t.Fatalf("normal enemy base mismatch")
	}
	g.difficulty = diffHard
	if g.enemyBaseByDifficulty() != 4 {
		t.Fatalf("hard enemy base mismatch")
	}
}

func TestEnemyMultipliersByDifficulty(t *testing.T) {
	g := newGame()
	g.difficulty = diffEasy
	if g.enemySpeedMultiplier() >= 1.0 || g.enemyFireBonus() >= 0 {
		t.Fatalf("easy should reduce speed and fire")
	}
	g.difficulty = diffHard
	if g.enemySpeedMultiplier() <= 1.0 || g.enemyHPBonus() <= 0 || g.enemyFireBonus() <= 0 {
		t.Fatalf("hard should increase speed/hp/fire")
	}
}

func TestSpawnWaveCountLowerBound(t *testing.T) {
	g := newPlayingGameForTest()
	g.enemies = nil
	g.enemyBase = -20
	g.spawnWave(1)
	if len(g.enemies) != enemyWaveMin {
		t.Fatalf("spawn lower bound expected %d, got %d", enemyWaveMin, len(g.enemies))
	}
}

func TestSpawnWaveCountUpperBound(t *testing.T) {
	g := newPlayingGameForTest()
	g.enemies = nil
	g.enemyBase = 20
	g.spawnWave(20)
	if len(g.enemies) != enemyWaveMax {
		t.Fatalf("spawn upper bound expected %d, got %d", enemyWaveMax, len(g.enemies))
	}
}

func TestDistributedSpawnsHasMultipleDirections(t *testing.T) {
	g := newGame()
	spawns := g.distributedSpawns()
	hasDown, hasLeft, hasRight := false, false, false
	for _, s := range spawns {
		switch s.dir {
		case down:
			hasDown = true
		case left:
			hasLeft = true
		case right:
			hasRight = true
		}
	}
	if !(hasDown && hasLeft && hasRight) {
		t.Fatalf("spawn directions should include down/left/right")
	}
}

func TestDifficultyAffectsEnemyStats(t *testing.T) {
	rand.Seed(1)
	easy := newGame()
	easy.difficulty = diffEasy
	easy.enemyBase = 3
	easy.startMatch()
	easy.enemies = nil
	easy.spawnWave(3)

	rand.Seed(1)
	hard := newGame()
	hard.difficulty = diffHard
	hard.enemyBase = 3
	hard.startMatch()
	hard.enemies = nil
	hard.spawnWave(3)

	if len(easy.enemies) == 0 || len(hard.enemies) == 0 {
		t.Fatalf("expected spawned enemies for both difficulties")
	}
	if hard.enemies[0].hp <= easy.enemies[0].hp {
		t.Fatalf("hard mode should have higher enemy hp")
	}
	if hard.enemies[0].speed <= easy.enemies[0].speed {
		t.Fatalf("hard mode should have higher enemy speed")
	}
}

func TestSpawnWaveEnemiesDoNotOverlapPlayerOrEachOther(t *testing.T) {
	rand.Seed(2)
	g := newPlayingGameForTest()
	g.enemies = nil
	g.enemyBase = 8
	g.spawnWave(3)

	for i := 0; i < len(g.enemies); i++ {
		er := tankRect(*g.enemies[i])
		if overlap(er, tankRect(g.player)) {
			t.Fatalf("enemy %d overlaps player at spawn", i)
		}
		for j := i + 1; j < len(g.enemies); j++ {
			if overlap(er, tankRect(*g.enemies[j])) {
				t.Fatalf("enemies %d and %d overlap at spawn", i, j)
			}
		}
	}
}

func TestAddDestructibleChunksCreatesOnlySquares(t *testing.T) {
	g := newGame()
	g.walls = nil
	g.addDestructibleChunks(rect{x: 10, y: 10, w: 20, h: 34}, tankSize, 1, true)
	if len(g.walls) == 0 {
		t.Fatalf("expected square chunks to be created")
	}
	for _, w := range g.walls {
		if !w.destructive {
			t.Fatalf("chunk should be destructive")
		}
		if w.box.w != tankSize || w.box.h != tankSize {
			t.Fatalf("expected full square chunk %.2fx%.2f, got %.2fx%.2f", tankSize, tankSize, w.box.w, w.box.h)
		}
	}
}

func TestArenaObstacleChunksMatchTankWidth(t *testing.T) {
	g := newGame()
	g.startMatch()
	found := false
	for _, w := range g.walls {
		if !w.destructive || w.guard {
			continue
		}
		found = true
		if w.box.w != tankSize {
			t.Fatalf("obstacle chunk width should match tank width %.2f, got %.2f", tankSize, w.box.w)
		}
	}
	if !found {
		t.Fatalf("expected destructible obstacle chunks")
	}
}

func TestScreenAndFortressAlignToGrid(t *testing.T) {
	g := newPlayingGameForTest()
	if screenW%gridSize != 0 || screenH%gridSize != 0 {
		t.Fatalf("screen size must be integer multiple of grid: %dx%d grid=%d", screenW, screenH, gridSize)
	}
	if int(g.fort.box.x)%gridSize != 0 || int(g.fort.box.y)%gridSize != 0 {
		t.Fatalf("fortress must align to grid lines, got x=%.2f y=%.2f", g.fort.box.x, g.fort.box.y)
	}
	bottomGap := screenH - int(g.fort.box.y+g.fort.box.h)
	if bottomGap != gridSize {
		t.Fatalf("fortress bottom gap should be one grid (%d), got %d", gridSize, bottomGap)
	}
}

func TestPlayerSpawnIsNotOverlappedAndCanMove(t *testing.T) {
	g := newPlayingGameForTest()
	pr := tankRect(g.player)
	if overlap(pr, g.fort.box) {
		t.Fatalf("player should not overlap fortress at spawn")
	}
	for _, w := range g.walls {
		if overlap(pr, w.box) {
			t.Fatalf("player should not overlap wall at spawn")
		}
	}

	startY := g.player.y
	if !g.tryMoveTank(&g.player, 0, -g.player.speed) {
		t.Fatalf("player should be able to move upward from spawn")
	}
	if g.player.y >= startY {
		t.Fatalf("player Y should decrease after upward move")
	}
}

func TestArenaObstacleLayoutStableWithSameSeed(t *testing.T) {
	_ = newGame()
	rand.Seed(101)
	g1 := newGame()
	g1.startMatch()
	sig1 := obstacleSignature(g1.walls)

	rand.Seed(101)
	g2 := newGame()
	g2.startMatch()
	sig2 := obstacleSignature(g2.walls)

	if sig1 != sig2 {
		t.Fatalf("same seed should produce same obstacle layout")
	}
}

func TestArenaObstacleLayoutVariesAcrossSeeds(t *testing.T) {
	_ = newGame()
	rand.Seed(102)
	g1 := newGame()
	g1.startMatch()
	sig1 := obstacleSignature(g1.walls)

	rand.Seed(103)
	g2 := newGame()
	g2.startMatch()
	sig2 := obstacleSignature(g2.walls)

	if sig1 == sig2 {
		t.Fatalf("different seeds should produce different obstacle layouts")
	}
}

func obstacleSignature(walls []*wall) string {
	var b strings.Builder
	for _, w := range walls {
		if !w.destructive || w.guard {
			continue
		}
		b.WriteString(fmt.Sprintf("%.0f:%.0f|", w.box.x, w.box.y))
	}
	return b.String()
}
