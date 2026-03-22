package tankbattle

const (
	gridSize = 30

	screenGridW = 32
	screenGridH = 22

	screenW = gridSize * screenGridW
	screenH = gridSize * screenGridH

	tankSize   = 34.0
	bulletSize = 6.0

	playerHullMaxHP   = 5
	playerTurretMaxHP = 5

	menuItemCount = 5

	enemyBaseMin = 1
	enemyBaseMax = 8
	enemyWaveMin = 1
	enemyWaveMax = 10
	matchWaveMin = 1
	matchWaveMax = 5

	powerupMaxActive = 3
	powerupSize      = 16.0

	fortressMaxHP = 10
	guardHitLoss  = 2
	fortHitLoss   = 3
	fortHitDamage = 3

	playerTurnDoubleTapFrames = 16
	playerTapGraceFrames      = 5
	playerTurnMoveLockFrames  = 3

	playerFireCooldownFrames      = 13
	playerRapidFireCooldownFrames = 7

	enemyFireCooldownBaseMin = 28
	enemyFireCooldownBaseVar = 36
	enemyFireCooldownMin     = 20

	scoreHistoryLimit = 100
	hudRankingRows    = 10
	hudRankingLineGap = 26
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
	speed       float64
	cooldown    int
	maxHP       int
	turretHP    int
	turretMaxHP int
	turnLock    int
	turnWant    direction
	turnVote    int
	hp          int
	isPlayer    bool
	aiTick      int
	age         int
	aiRand      float64
	replan      int
	fireBias    int
	aggro       float64
	role        enemyRole
	stuck       int
}

type bullet struct {
	x          float64
	y          float64
	vx         float64
	vy         float64
	fromPlayer bool
	alive      bool
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

type spawnPoint struct {
	x   float64
	y   float64
	dir direction
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

	score              int
	enemyKills         int
	win                bool
	paused             bool
	frame              int
	audioFrame         int
	wave               int
	maxWave            int
	waveDelay          int
	msg                string
	msgTick            int
	shieldTick         int
	rapidTick          int
	playerSilentFrames int

	difficulty   difficulty
	enemyBase    int
	totalWaves   int
	menuIndex    int
	soundEnabled bool
	soundVolume  int
	audio        sfxPlayer
	scoreHistory []scoreEntry
	rankScroll   int
	matchLogged  bool
	showHistory  bool

	menuResumeAvailable bool
	menuReturnState     gameState
	menuReturnPaused    bool
	menuRequireRestart  bool

	playerTapFrame      [4]int
	playerPressStart    [4]int
	playerMoveLockUntil int
}

type sfxPlayer interface {
	Play(id sfxID, frame int)
	SetEnabled(enabled bool)
	SetSFXVolume(volume float64)
	Enabled() bool
}
