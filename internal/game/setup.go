package game

import (
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"
)

var randSeedOnce sync.Once

type newGameOptions struct {
	loadUserSettings bool
	persistUserData  bool
	audio            sfxPlayer
	debug            *DebugController
	randomSeed       *int64
}

func newGame() *game {
	return newGameWithOptions(newGameOptions{
		loadUserSettings: true,
		persistUserData:  true,
		audio:            newAudioManager(),
	})
}

func newGameWithOptions(opts newGameOptions) *game {
	seedGameplayRandom(opts.randomSeed)
	g := &game{
		state:           stateMenu,
		difficulty:      diffNormal,
		enemyBase:       3,
		totalWaves:      4,
		menuIndex:       0,
		soundEnabled:    true,
		soundVolume:     75,
		persistUserData: opts.persistUserData,
		audio:           opts.audio,
		debug:           opts.debug,
	}
	if opts.loadUserSettings {
		g.loadUserSettings()
	}
	return g
}

func seedGameplayRandom(seed *int64) {
	if seed != nil {
		rand.Seed(*seed)
		return
	}
	randSeedOnce.Do(func() {
		now := time.Now()
		randomSeed := now.UnixNano() ^ (int64(os.Getpid()) << 17) ^ now.UnixMicro()
		rand.Seed(randomSeed)
	})
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
	g.addDestructibleChunks(rect{x: bx - block, y: by + block, w: bw + block*2, h: block}, block, guardHP, true)
}

func (g *game) buildArenaObstacles() {
	for _, box := range g.pickArenaObstacleLayout() {
		g.addDestructibleChunks(box, tankSize, 1, false)
	}
	g.walls = append(g.walls,
		&wall{box: rect{40, 320, 80, 16}, hp: 99, maxHP: 99, destructive: false},
		&wall{box: rect{840, 320, 80, 16}, hp: 99, maxHP: 99, destructive: false},
	)
}

func (g *game) pickArenaObstacleLayout() []rect {
	baseTemplates := [][]rect{
		{
			{150, 160, 140, 26},
			{350, 285, 260, 24},
			{690, 160, 140, 26},
			{170, 450, 220, 24},
			{590, 450, 210, 24},
		},
		{
			{130, 170, 170, 24},
			{330, 300, 300, 22},
			{670, 170, 170, 24},
			{180, 430, 190, 24},
			{610, 430, 190, 24},
		},
		{
			{160, 150, 120, 26},
			{320, 275, 320, 24},
			{700, 150, 120, 26},
			{150, 470, 230, 22},
			{580, 470, 230, 22},
		},
	}
	pick := rand.Intn(len(baseTemplates))
	for attempt := 0; attempt < 12; attempt++ {
		boxes := jitterObstacleLayout(baseTemplates[pick], gridSize)
		if g.isArenaObstacleLayoutSafe(boxes) {
			return boxes
		}
		pick = rand.Intn(len(baseTemplates))
	}
	return append([]rect(nil), baseTemplates[0]...)
}

func jitterObstacleLayout(base []rect, step int) []rect {
	if step <= 0 {
		return append([]rect(nil), base...)
	}
	out := make([]rect, 0, len(base))
	for i, b := range base {
		dx := float64(jitterStep(step))
		dy := float64(jitterStep(step))
		// Keep the center lane stable to avoid accidental hard locks.
		if i == 1 {
			dx = 0
		}
		box := rect{x: b.x + dx, y: b.y + dy, w: b.w, h: b.h}
		if i == 3 || i == 4 {
			// Bottom lane should not drift toward the fortress.
			maxY := float64(screenH - gridSize*7)
			if box.y > maxY {
				box.y = maxY
			}
		}
		out = append(out, box)
	}
	return out
}

func jitterStep(step int) int {
	switch rand.Intn(3) {
	case 0:
		return -step
	case 1:
		return 0
	default:
		return step
	}
}

func (g *game) isArenaObstacleLayoutSafe(boxes []rect) bool {
	playerSafe := rect{
		x: g.player.x - float64(gridSize*2),
		y: g.player.y - float64(gridSize),
		w: tankSize + float64(gridSize*4),
		h: tankSize + float64(gridSize*3),
	}
	fortSafe := rect{
		x: g.fort.box.x - float64(gridSize*3),
		y: g.fort.box.y - float64(gridSize*3),
		w: g.fort.box.w + float64(gridSize*6),
		h: g.fort.box.h + float64(gridSize*5),
	}
	for _, b := range boxes {
		if b.x < 0 || b.y < 0 || b.x+b.w > screenW || b.y+b.h > screenH {
			return false
		}
		if overlap(b, playerSafe) || overlap(b, fortSafe) {
			return false
		}
	}
	return true
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
