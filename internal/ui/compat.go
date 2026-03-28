package ui

import (
	"image/color"
	"strings"
	"sync"
)

const (
	gridSize          = GridSize
	screenW           = ScreenW
	screenH           = ScreenH
	tankSize          = TankSize
	bulletSize        = BulletSize
	menuItemCount     = MenuItemCount
	matchWaveMin      = MatchWaveMin
	matchWaveMax      = MatchWaveMax
	hudRankingRows    = HUDRankingRows
	hudRankingLineGap = HUDRankingLineGap
)

type direction int

const (
	up direction = iota
	down
	left
	right
)

type gameState int

const (
	stateMenu gameState = iota
	statePlaying
	stateEnded
)

type difficulty int

const (
	diffEasy difficulty = iota
	diffNormal
	diffHard
)

type enemyRole int

const (
	roleAssault enemyRole = iota
	roleLeftFlank
	roleRightFlank
)

type powerupKind int

const (
	powerShield powerupKind = iota
	powerRapid
	powerRepair
)

type rect struct {
	x float64
	y float64
	w float64
	h float64
}

type wall struct {
	box         rect
	hp          int
	maxHP       int
	destructive bool
	guard       bool
}

type fortress struct {
	box   rect
	hp    int
	maxHP int
}

type tank struct {
	x           float64
	y           float64
	dir         direction
	turret      direction
	maxHP       int
	turretHP    int
	turretMaxHP int
	hp          int
	role        enemyRole
}

type bullet struct {
	x          float64
	y          float64
	vx         float64
	vy         float64
	fromPlayer bool
	dmg        int
}

type explosion struct {
	x      float64
	y      float64
	radius float64
	life   int
	max    int
}

type powerup struct {
	kind powerupKind
	box  rect
	life int
}

type scoreEntry struct {
	Score       int
	At          string
	DurationSec int
}

type game struct {
	state gameState

	player     tank
	enemies    []tank
	bullets    []bullet
	walls      []wall
	fort       fortress
	explosions []explosion
	powerups   []powerup

	score               int
	win                 bool
	paused              bool
	audioFrame          int
	wave                int
	maxWave             int
	msg                 string
	shieldTick          int
	rapidTick           int
	difficulty          difficulty
	totalWaves          int
	menuIndex           int
	soundEnabled        bool
	soundVolume         int
	scoreHistory        []scoreEntry
	rankScroll          int
	showHistory         bool
	menuResumeAvailable bool
	menuRequireRestart  bool
	backgroundSeed      int64
	matchIntroTick      int
	matchIntroMax       int

	bestScoreValue   int
	currentRankValue int
}

func newCompatGame(snapshot Snapshot) *game {
	g := compatGamePool.Get().(*game)
	fillCompatGame(g, snapshot)
	return g
}

func releaseCompatGame(g *game) {
	if g == nil {
		return
	}
	compatGamePool.Put(g)
}

var compatGamePool = sync.Pool{
	New: func() any {
		return &game{}
	},
}

func fillCompatGame(g *game, snapshot Snapshot) {
	g.state = snapshotState(snapshot.State)
	g.player = snapshotTank(snapshot.Player)
	g.fort = snapshotFortress(snapshot.Fort)
	g.score = snapshot.Score
	g.win = snapshot.Win
	g.paused = snapshot.Paused
	g.audioFrame = snapshot.AudioFrame
	g.wave = snapshot.Wave
	g.maxWave = snapshot.MaxWave
	g.msg = snapshot.Message
	g.shieldTick = snapshot.ShieldTick
	g.rapidTick = snapshot.RapidTick
	g.difficulty = snapshotDifficulty(snapshot.Difficulty)
	g.totalWaves = snapshot.TotalWaves
	g.menuIndex = snapshot.MenuIndex
	g.soundEnabled = snapshot.SoundEnabled
	g.soundVolume = snapshot.SoundVolume
	g.rankScroll = snapshot.RankScroll
	g.showHistory = snapshot.ShowHistory
	g.menuResumeAvailable = snapshot.MenuResumeAvailable
	g.menuRequireRestart = snapshot.MenuRequireRestart
	g.bestScoreValue = snapshot.BestScore
	g.currentRankValue = snapshot.CurrentRank
	g.backgroundSeed = snapshot.BackgroundSeed
	g.matchIntroTick = snapshot.MatchIntroTick
	g.matchIntroMax = snapshot.MatchIntroMax

	g.enemies = g.enemies[:0]
	for _, enemy := range snapshot.Enemies {
		g.enemies = append(g.enemies, snapshotTank(enemy))
	}

	g.bullets = g.bullets[:0]
	for _, shot := range snapshot.Bullets {
		g.bullets = append(g.bullets, snapshotBullet(shot))
	}

	g.walls = g.walls[:0]
	for _, item := range snapshot.Walls {
		g.walls = append(g.walls, snapshotWall(item))
	}

	g.explosions = g.explosions[:0]
	for _, item := range snapshot.Explosions {
		g.explosions = append(g.explosions, snapshotExplosion(item))
	}

	g.powerups = g.powerups[:0]
	for _, item := range snapshot.Powerups {
		g.powerups = append(g.powerups, snapshotPowerup(item))
	}

	g.scoreHistory = g.scoreHistory[:0]
	for _, entry := range snapshot.ScoreHistory {
		g.scoreHistory = append(g.scoreHistory, scoreEntry(entry))
	}

	g.clampRankScroll()
}

func snapshotState(value string) gameState {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case StatePlaying:
		return statePlaying
	case StateEnded:
		return stateEnded
	default:
		return stateMenu
	}
}

func snapshotDifficulty(value string) difficulty {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "easy":
		return diffEasy
	case "hard":
		return diffHard
	default:
		return diffNormal
	}
}

func snapshotDirection(value Direction) direction {
	switch value {
	case Down:
		return down
	case Left:
		return left
	case Right:
		return right
	default:
		return up
	}
}

func snapshotRole(value Role) enemyRole {
	switch value {
	case RoleLeftFlank:
		return roleLeftFlank
	case RoleRightFlank:
		return roleRightFlank
	default:
		return roleAssault
	}
}

func snapshotPowerupKind(value PowerupKind) powerupKind {
	switch value {
	case PowerRapid:
		return powerRapid
	case PowerRepair:
		return powerRepair
	default:
		return powerShield
	}
}

func snapshotRect(value Rect) rect {
	return rect{x: value.X, y: value.Y, w: value.W, h: value.H}
}

func snapshotWall(value Wall) wall {
	return wall{
		box:         snapshotRect(value.Box),
		hp:          value.HP,
		maxHP:       value.MaxHP,
		destructive: value.Destructive,
		guard:       value.Guard,
	}
}

func snapshotFortress(value Fortress) fortress {
	return fortress{
		box:   snapshotRect(value.Box),
		hp:    value.HP,
		maxHP: value.MaxHP,
	}
}

func snapshotTank(value Tank) tank {
	return tank{
		x:           value.X,
		y:           value.Y,
		dir:         snapshotDirection(value.Dir),
		turret:      snapshotDirection(value.Turret),
		maxHP:       value.MaxHP,
		turretHP:    value.TurretHP,
		turretMaxHP: value.TurretMaxHP,
		hp:          value.HP,
		role:        snapshotRole(value.Role),
	}
}

func snapshotBullet(value Bullet) bullet {
	return bullet{
		x:          value.X,
		y:          value.Y,
		vx:         value.VX,
		vy:         value.VY,
		fromPlayer: value.FromPlayer,
		dmg:        value.Dmg,
	}
}

func snapshotExplosion(value Explosion) explosion {
	return explosion{
		x:      value.X,
		y:      value.Y,
		radius: value.Radius,
		life:   value.Life,
		max:    value.Max,
	}
}

func snapshotPowerup(value Powerup) powerup {
	return powerup{
		kind: snapshotPowerupKind(value.Kind),
		box:  snapshotRect(value.Box),
		life: value.Life,
	}
}

func (g *game) bestScore() int {
	if g.bestScoreValue > 0 {
		return g.bestScoreValue
	}
	best := 0
	for _, entry := range g.scoreHistory {
		if entry.Score > best {
			best = entry.Score
		}
	}
	if g.score > best {
		best = g.score
	}
	return best
}

func (g *game) currentRank() int {
	if g.currentRankValue > 0 {
		return g.currentRankValue
	}
	rank := 1
	for _, entry := range g.scoreHistory {
		if entry.Score > g.score {
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

func toRGBA(c color.Color) color.RGBA {
	r, g, b, a := c.RGBA()
	return color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
}

func shift(c color.RGBA, dr, dg, db int) color.RGBA {
	r := clampInt(int(c.R)+dr, 0, 255)
	g := clampInt(int(c.G)+dg, 0, 255)
	b := clampInt(int(c.B)+db, 0, 255)
	return color.RGBA{uint8(r), uint8(g), uint8(b), c.A}
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func onOffText(on bool) string {
	if on {
		return "ON"
	}
	return "OFF"
}
