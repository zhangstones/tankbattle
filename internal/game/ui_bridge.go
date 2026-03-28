package game

import (
	gameui "tankbattle/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
)

func (g *game) Draw(screen *ebiten.Image) {
	gameui.Draw(screen, g.uiSnapshot())
}

func (g *game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return gameui.Layout(outsideWidth, outsideHeight)
}

func (g *game) uiSnapshot() gameui.Snapshot {
	if g == nil {
		return gameui.Snapshot{}
	}

	snapshot := gameui.Snapshot{
		State:               debugStateName(g.state),
		Player:              uiTank(g.player),
		Fort:                uiFortress(g.fort),
		Score:               g.score,
		Win:                 g.win,
		Paused:              g.paused,
		AudioFrame:          g.audioFrame,
		Wave:                g.wave,
		MaxWave:             g.maxWave,
		Message:             g.msg,
		ShieldTick:          g.shieldTick,
		RapidTick:           g.rapidTick,
		Difficulty:          debugDifficultyName(g.difficulty),
		TotalWaves:          g.totalWaves,
		MenuIndex:           g.menuIndex,
		SoundEnabled:        g.soundEnabled,
		SoundVolume:         g.soundVolume,
		RankScroll:          g.rankScroll,
		ShowHistory:         g.showHistory,
		MenuResumeAvailable: g.menuResumeAvailable,
		MenuRequireRestart:  g.menuRequireRestart,
		BestScore:           g.bestScore(),
		CurrentRank:         g.currentRank(),
		BackgroundSeed:      g.backgroundSeed,
		MatchIntroTick:      g.matchIntroTick,
		MatchIntroMax:       g.matchIntroMax,
	}

	snapshot.Enemies = make([]gameui.Tank, 0, len(g.enemies))
	for _, enemy := range g.enemies {
		if enemy == nil {
			continue
		}
		snapshot.Enemies = append(snapshot.Enemies, uiTank(*enemy))
	}

	snapshot.Bullets = make([]gameui.Bullet, 0, len(g.bullets))
	for _, shot := range g.bullets {
		if shot == nil {
			continue
		}
		snapshot.Bullets = append(snapshot.Bullets, uiBullet(*shot))
	}

	snapshot.Walls = make([]gameui.Wall, 0, len(g.walls))
	for _, item := range g.walls {
		if item == nil {
			continue
		}
		snapshot.Walls = append(snapshot.Walls, uiWall(*item))
	}

	snapshot.Explosions = make([]gameui.Explosion, 0, len(g.explosions))
	for _, item := range g.explosions {
		if item == nil {
			continue
		}
		snapshot.Explosions = append(snapshot.Explosions, uiExplosion(*item))
	}

	snapshot.Powerups = make([]gameui.Powerup, 0, len(g.powerups))
	for _, item := range g.powerups {
		if item == nil {
			continue
		}
		snapshot.Powerups = append(snapshot.Powerups, uiPowerup(*item))
	}

	snapshot.ScoreHistory = make([]gameui.ScoreEntry, 0, len(g.scoreHistory))
	for _, entry := range g.scoreHistory {
		snapshot.ScoreHistory = append(snapshot.ScoreHistory, gameui.ScoreEntry(entry))
	}

	return snapshot
}

func uiRect(value rect) gameui.Rect {
	return gameui.Rect{X: value.x, Y: value.y, W: value.w, H: value.h}
}

func uiWall(value wall) gameui.Wall {
	return gameui.Wall{
		Box:         uiRect(value.box),
		HP:          value.hp,
		MaxHP:       value.maxHP,
		Destructive: value.destructive,
		Guard:       value.guard,
	}
}

func uiFortress(value fortress) gameui.Fortress {
	return gameui.Fortress{
		Box:   uiRect(value.box),
		HP:    value.hp,
		MaxHP: value.maxHP,
	}
}

func uiTank(value tank) gameui.Tank {
	return gameui.Tank{
		X:           value.x,
		Y:           value.y,
		Dir:         uiDirection(value.dir),
		Turret:      uiDirection(value.turret),
		MaxHP:       value.maxHP,
		TurretHP:    value.turretHP,
		TurretMaxHP: value.turretMaxHP,
		HP:          value.hp,
		Role:        uiRole(value.role),
	}
}

func uiBullet(value bullet) gameui.Bullet {
	return gameui.Bullet{
		X:          value.x,
		Y:          value.y,
		VX:         value.vx,
		VY:         value.vy,
		FromPlayer: value.fromPlayer,
		Dmg:        value.dmg,
	}
}

func uiExplosion(value explosion) gameui.Explosion {
	return gameui.Explosion{
		X:      value.x,
		Y:      value.y,
		Radius: value.radius,
		Life:   value.life,
		Max:    value.max,
	}
}

func uiPowerup(value powerup) gameui.Powerup {
	return gameui.Powerup{
		Kind: uiPowerupKind(value.kind),
		Box:  uiRect(value.box),
		Life: value.life,
	}
}

func uiDirection(value direction) gameui.Direction {
	switch value {
	case down:
		return gameui.Down
	case left:
		return gameui.Left
	case right:
		return gameui.Right
	default:
		return gameui.Up
	}
}

func uiRole(value enemyRole) gameui.Role {
	switch value {
	case roleLeftFlank:
		return gameui.RoleLeftFlank
	case roleRightFlank:
		return gameui.RoleRightFlank
	default:
		return gameui.RoleAssault
	}
}

func uiPowerupKind(value powerupKind) gameui.PowerupKind {
	switch value {
	case powerRapid:
		return gameui.PowerRapid
	case powerRepair:
		return gameui.PowerRepair
	default:
		return gameui.PowerShield
	}
}
