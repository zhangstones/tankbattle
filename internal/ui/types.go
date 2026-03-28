package ui

type Direction int

const (
	Up Direction = iota
	Down
	Left
	Right
)

type Role int

const (
	RoleAssault Role = iota
	RoleLeftFlank
	RoleRightFlank
)

type PowerupKind int

const (
	PowerShield PowerupKind = iota
	PowerRapid
	PowerRepair
)

type Rect struct {
	X float64
	Y float64
	W float64
	H float64
}

type Wall struct {
	Box         Rect
	HP          int
	MaxHP       int
	Destructive bool
	Guard       bool
}

type Fortress struct {
	Box   Rect
	HP    int
	MaxHP int
}

type Tank struct {
	X           float64
	Y           float64
	Dir         Direction
	Turret      Direction
	MaxHP       int
	TurretHP    int
	TurretMaxHP int
	HP          int
	Role        Role
}

type Bullet struct {
	X          float64
	Y          float64
	VX         float64
	VY         float64
	FromPlayer bool
	Dmg        int
}

type Explosion struct {
	X      float64
	Y      float64
	Radius float64
	Life   int
	Max    int
}

type Powerup struct {
	Kind PowerupKind
	Box  Rect
	Life int
}

type ScoreEntry struct {
	Score       int
	At          string
	DurationSec int
}

type Snapshot struct {
	State               string
	Player              Tank
	Enemies             []Tank
	Bullets             []Bullet
	Walls               []Wall
	Fort                Fortress
	Explosions          []Explosion
	Powerups            []Powerup
	Score               int
	Win                 bool
	Paused              bool
	AudioFrame          int
	Wave                int
	MaxWave             int
	Message             string
	ShieldTick          int
	RapidTick           int
	Difficulty          string
	TotalWaves          int
	MenuIndex           int
	SoundEnabled        bool
	SoundVolume         int
	ScoreHistory        []ScoreEntry
	RankScroll          int
	ShowHistory         bool
	MenuResumeAvailable bool
	MenuRequireRestart  bool
	BestScore           int
	CurrentRank         int
	BackgroundSeed      int64
	MatchIntroTick      int
	MatchIntroMax       int
}

const (
	StateMenu    = "menu"
	StatePlaying = "playing"
	StateEnded   = "ended"

	GridSize    = 30
	ScreenGridW = 32
	ScreenGridH = 22
	ScreenW     = GridSize * ScreenGridW
	ScreenH     = GridSize * ScreenGridH
	TankSize    = 34.0
	BulletSize  = 6.0

	MenuItemCount     = 5
	MatchWaveMin      = 1
	MatchWaveMax      = 5
	HUDRankingRows    = 10
	HUDRankingLineGap = 26
)
