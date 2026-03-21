package tankbattle

import (
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"
)

var randSeedOnce sync.Once

func newGame() *game {
	randSeedOnce.Do(func() {
		now := time.Now()
		seed := now.UnixNano() ^ (int64(os.Getpid()) << 17) ^ now.UnixMicro()
		rand.Seed(seed)
	})
	g := &game{
		state:        stateMenu,
		difficulty:   diffNormal,
		enemyBase:    3,
		totalWaves:   4,
		menuIndex:    0,
		soundEnabled: true,
		soundVolume:  75,
	}
	g.audio = newAudioManager()
	g.loadUserSettings()
	return g
}

func (g *game) startMatch() {
	g.state = statePlaying
	g.wave = 1
	g.maxWave = g.clampedTotalWaves()
	g.enemyBase = g.enemyBaseByDifficulty()
	g.score = 0
	g.enemyKills = 0
	g.win = false
	g.paused = false
	g.frame = 0
	g.waveDelay = 0
	g.msg = ""
	g.msgTick = 0
	g.shieldTick = 0
	g.rapidTick = 0
	g.playerSilentFrames = 0
	g.matchLogged = false
	g.rankScroll = 0
	g.showHistory = false
	g.menuResumeAvailable = false
	g.menuRequireRestart = false

	fortW := float64(gridSize * 2)
	fortH := float64(gridSize)
	fortX := float64((screenW - int(fortW)) / 2)
	fortY := float64(screenH - gridSize - int(fortH))
	g.fort = fortress{box: rect{x: fortX, y: fortY, w: fortW, h: fortH}, hp: fortressMaxHP, maxHP: fortressMaxHP}

	g.player = tank{
		x:           screenW/2 - tankSize/2,
		y:           g.fort.box.y - float64(gridSize) - tankSize,
		dir:         up,
		turret:      up,
		speed:       3.2,
		hp:          playerHullMaxHP,
		maxHP:       playerHullMaxHP,
		turretHP:    playerTurretMaxHP,
		turretMaxHP: playerTurretMaxHP,
		isPlayer:    true,
	}
	g.resetPlayerTapFrames()
	g.bullets = g.bullets[:0]
	g.explosions = g.explosions[:0]
	g.powerups = g.powerups[:0]
	g.enemies = g.enemies[:0]

	g.walls = make([]*wall, 0, 128)
	g.buildFortDefense()
	g.buildArenaObstacles()
	g.spawnWave(g.wave)
	g.setMessage(fmt.Sprintf("Wave %d incoming", g.wave), 120)
}

func (g *game) resetPlayerTapFrames() {
	for i := range g.playerTapFrame {
		g.playerTapFrame[i] = -9999
		g.playerPressStart[i] = -9999
	}
	g.playerMoveLockUntil = 0
}

func (g *game) maxWaveByDifficulty() int {
	switch g.difficulty {
	case diffEasy:
		return 3
	case diffHard:
		return 5
	default:
		return 4
	}
}

func (g *game) enemyBaseByDifficulty() int {
	switch g.difficulty {
	case diffEasy:
		return 2
	case diffHard:
		return 4
	default:
		return 3
	}
}

func (g *game) clampedTotalWaves() int {
	if g.totalWaves < matchWaveMin {
		return matchWaveMin
	}
	if g.totalWaves > matchWaveMax {
		return matchWaveMax
	}
	return g.totalWaves
}

func (g *game) enemySpeedMultiplier() float64 {
	switch g.difficulty {
	case diffEasy:
		return 0.92
	case diffHard:
		return 1.16
	default:
		return 1.0
	}
}

func (g *game) enemyHPBonus() int {
	switch g.difficulty {
	case diffHard:
		return 1
	default:
		return 0
	}
}

func (g *game) enemyFireBonus() int {
	switch g.difficulty {
	case diffEasy:
		return -6
	case diffHard:
		return 8
	default:
		return 0
	}
}

func (g *game) buildFortDefense() {
	bx := g.fort.box.x
	by := g.fort.box.y
	bw := g.fort.box.w
	block := float64(gridSize)
	guardHP := 2
	g.addDestructibleChunks(rect{x: bx - block, y: by - block, w: bw + block*2, h: block}, block, guardHP, true)
	g.addDestructibleChunks(rect{x: bx - block, y: by, w: block, h: block}, block, guardHP, true)
	g.addDestructibleChunks(rect{x: bx + bw, y: by, w: block, h: block}, block, guardHP, true)
}

func (g *game) buildArenaObstacles() {
	g.addDestructibleChunks(rect{150, 160, 140, 26}, tankSize, 1, false)
	g.addDestructibleChunks(rect{350, 285, 260, 24}, tankSize, 1, false)
	g.addDestructibleChunks(rect{690, 160, 140, 26}, tankSize, 1, false)
	g.addDestructibleChunks(rect{170, 450, 220, 24}, tankSize, 1, false)
	g.addDestructibleChunks(rect{590, 450, 210, 24}, tankSize, 1, false)
	g.walls = append(g.walls,
		&wall{box: rect{40, 320, 80, 16}, hp: 99, maxHP: 99, destructive: false},
		&wall{box: rect{840, 320, 80, 16}, hp: 99, maxHP: 99, destructive: false},
	)
}

func (g *game) distributedSpawns() []spawnPoint {
	return []spawnPoint{
		{80, 58, down}, {190, 58, down}, {screenW/2 - tankSize/2, 58, down}, {screenW - 220, 58, down}, {screenW - 110, 58, down},
		{20, 150, right}, {20, 270, right}, {20, 390, right}, {20, 510, right},
		{screenW - tankSize - 20, 150, left}, {screenW - tankSize - 20, 270, left}, {screenW - tankSize - 20, 390, left}, {screenW - tankSize - 20, 510, left},
	}
}

func (g *game) spawnWave(wave int) {
	count := g.enemyBase + wave - 1
	if count < enemyWaveMin {
		count = enemyWaveMin
	}
	if count > enemyWaveMax {
		count = enemyWaveMax
	}
	spawns := g.distributedSpawns()
	perm := rand.Perm(len(spawns))

	for i := 0; i < count; i++ {
		p := spawns[perm[i%len(spawns)]]
		hp := 1 + wave/2 + g.enemyHPBonus()
		if hp < 1 {
			hp = 1
		}
		e := &tank{
			x:      p.x,
			y:      p.y,
			dir:    p.dir,
			turret: p.dir,
			speed:  (1.55 + float64(wave)*0.2) * g.enemySpeedMultiplier(),
			hp:     hp,
			maxHP:  hp,
			aiTick: rand.Intn(20) + 8,
			role:   enemyRole(i % 3),
		}
		initEnemyTraits(e)
		if p.dir == right {
			e.role = roleLeftFlank
		} else if p.dir == left {
			e.role = roleRightFlank
		}
		if g.placeEnemy(e, p.x, p.y) {
			g.enemies = append(g.enemies, e)
		}
	}
}

func initEnemyTraits(e *tank) {
	e.aiRand = rand.Float64()
	e.replan = 5 + rand.Intn(7)
	e.fireBias = rand.Intn(11) - 5
	e.aggro = 0.85 + rand.Float64()*0.55
}

func (g *game) addDestructibleChunks(box rect, chunk float64, hp int, guard bool) {
	if chunk <= 0 {
		return
	}
	cols := int(box.w / chunk)
	rows := int(box.h / chunk)
	if cols < 1 {
		cols = 1
	}
	if rows < 1 {
		rows = 1
	}
	if cols <= 0 || rows <= 0 {
		return
	}
	offsetX := (box.w - float64(cols)*chunk) / 2
	offsetY := (box.h - float64(rows)*chunk) / 2
	startX := box.x + offsetX
	startY := box.y + offsetY
	for r := 0; r < rows; r++ {
		y := startY + float64(r)*chunk
		for c := 0; c < cols; c++ {
			x := startX + float64(c)*chunk
			g.walls = append(g.walls, &wall{
				box:         rect{x: x, y: y, w: chunk, h: chunk},
				hp:          hp,
				maxHP:       hp,
				destructive: true,
				guard:       guard,
			})
		}
	}
}
