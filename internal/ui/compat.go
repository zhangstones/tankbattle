package ui

import (
	"image/color"
	"strings"
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
	enemies    []*tank
	bullets    []*bullet
	walls      []*wall
	fort       fortress
	explosions []*explosion
	powerups   []*powerup

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

	bestScoreValue   int
	currentRankValue int
}

func newCompatGame(snapshot Snapshot) *game {
	g := &game{
		state:               snapshotState(snapshot.State),
		player:              snapshotTank(snapshot.Player),
		fort:                snapshotFortress(snapshot.Fort),
		score:               snapshot.Score,
		win:                 snapshot.Win,
		paused:              snapshot.Paused,
		audioFrame:          snapshot.AudioFrame,
		wave:                snapshot.Wave,
		maxWave:             snapshot.MaxWave,
		msg:                 snapshot.Message,
		shieldTick:          snapshot.ShieldTick,
		rapidTick:           snapshot.RapidTick,
		difficulty:          snapshotDifficulty(snapshot.Difficulty),
		totalWaves:          snapshot.TotalWaves,
		menuIndex:           snapshot.MenuIndex,
		soundEnabled:        snapshot.SoundEnabled,
		soundVolume:         snapshot.SoundVolume,
		rankScroll:          snapshot.RankScroll,
		showHistory:         snapshot.ShowHistory,
		menuResumeAvailable: snapshot.MenuResumeAvailable,
		menuRequireRestart:  snapshot.MenuRequireRestart,
		bestScoreValue:      snapshot.BestScore,
		currentRankValue:    snapshot.CurrentRank,
	}

	g.enemies = make([]*tank, 0, len(snapshot.Enemies))
	for _, enemy := range snapshot.Enemies {
		e := snapshotTank(enemy)
		g.enemies = append(g.enemies, &e)
	}

	g.bullets = make([]*bullet, 0, len(snapshot.Bullets))
	for _, shot := range snapshot.Bullets {
		b := snapshotBullet(shot)
		g.bullets = append(g.bullets, &b)
	}

	g.walls = make([]*wall, 0, len(snapshot.Walls))
	for _, item := range snapshot.Walls {
		w := snapshotWall(item)
		g.walls = append(g.walls, &w)
	}

	g.explosions = make([]*explosion, 0, len(snapshot.Explosions))
	for _, item := range snapshot.Explosions {
		ex := snapshotExplosion(item)
		g.explosions = append(g.explosions, &ex)
	}

	g.powerups = make([]*powerup, 0, len(snapshot.Powerups))
	for _, item := range snapshot.Powerups {
		p := snapshotPowerup(item)
		g.powerups = append(g.powerups, &p)
	}

	g.scoreHistory = make([]scoreEntry, 0, len(snapshot.ScoreHistory))
	for _, entry := range snapshot.ScoreHistory {
		g.scoreHistory = append(g.scoreHistory, scoreEntry(entry))
	}

	g.clampRankScroll()
	return g
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
