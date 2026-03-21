package tankbattle

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func (g *game) Update() error {
	g.audioFrame++

	if inpututil.IsKeyJustPressed(ebiten.KeyR) && g.restartIfAllowed() {
		return nil
	}

	switch g.state {
	case stateMenu:
		g.updateMenu()
		return nil
	case stateEnded:
		if inpututil.IsKeyJustPressed(ebiten.KeyH) {
			g.toggleHistoryView()
		}
		if g.showHistory {
			g.updateRankScrollInput()
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyM) {
			g.playSFX(sfxMenuConfirm)
			g.returnToMenu()
		}
		return nil
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyM) {
		g.playSFX(sfxMenuConfirm)
		g.returnToMenu()
		return nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyH) {
		g.toggleHistoryView()
	}
	if g.showHistory {
		g.updateRankScrollInput()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		g.togglePause()
	}
	if g.paused {
		return nil
	}

	g.frame++
	g.playerSilentFrames++
	if g.msgTick > 0 {
		g.msgTick--
		if g.msgTick == 0 {
			g.msg = ""
		}
	}
	if g.shieldTick > 0 {
		g.shieldTick--
		if g.shieldTick == 0 {
			g.playSFX(sfxBuffShieldOff)
		}
	}
	if g.rapidTick > 0 {
		g.rapidTick--
		if g.rapidTick == 0 {
			g.playSFX(sfxBuffRapidOff)
		}
	}

	g.updatePlayer()
	g.updateEnemies()
	g.updateBullets()
	g.updatePowerups()
	g.updateExplosions()
	g.cleanupWalls()
	g.trySpawnRandomPowerup()

	if g.fort.hp <= 0 || g.player.hp <= 0 {
		if g.player.hp <= 0 {
			g.playSFX(sfxDestroyPlayer)
		}
		g.applyDefeatEnergyState()
		g.finishMatch(false)
		g.playSFX(sfxLose)
		return nil
	}

	if len(g.enemies) == 0 {
		if g.waveDelay == 0 {
			if g.wave >= g.maxWave {
				g.finishMatch(true)
				g.playSFX(sfxWin)
			} else {
				g.wave++
				g.waveDelay = 130
				g.setMessage(fmt.Sprintf("Prepare wave %d", g.wave), g.waveDelay)
				g.playSFX(sfxWavePrepare)
			}
		} else {
			g.waveDelay--
			if g.waveDelay == 0 {
				g.spawnWave(g.wave)
				g.setMessage(fmt.Sprintf("Wave %d incoming", g.wave), 100)
				g.playSFX(sfxWaveStart)
			}
		}
	}
	return nil
}

func (g *game) finishMatch(win bool) {
	g.state = stateEnded
	g.win = win
	g.appendCurrentScoreHistory()
}

func (g *game) updateRankScrollInput() {
	scroll := 0
	if inpututil.IsKeyJustPressed(ebiten.KeyPageUp) {
		scroll--
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyPageDown) {
		scroll++
	}
	_, wheelY := ebiten.Wheel()
	if wheelY > 0.2 {
		scroll--
	} else if wheelY < -0.2 {
		scroll++
	}
	if scroll != 0 {
		g.rankScroll += scroll
		g.clampRankScroll()
	}
}

func (g *game) toggleHistoryView() {
	g.showHistory = !g.showHistory
	g.clampRankScroll()
}

func (g *game) restartIfAllowed() bool {
	if g.state == stateMenu {
		return false
	}
	g.startMatch()
	return true
}

func (g *game) returnToMenu() {
	g.state = stateMenu
	g.paused = false
	g.showHistory = false
}

func (g *game) togglePause() {
	g.paused = !g.paused
	g.playSFX(sfxPauseToggle)
	if g.paused {
		g.setMessage("Paused", 999999)
		return
	}
	g.setMessage("Resume", 45)
}

func (g *game) cleanupWalls() {
	keep := g.walls[:0]
	for _, w := range g.walls {
		if !w.destructive || w.hp > 0 {
			keep = append(keep, w)
		}
	}
	g.walls = keep
}

func (g *game) updateExplosions() {
	alive := g.explosions[:0]
	for _, ex := range g.explosions {
		ex.life--
		if ex.life > 0 {
			alive = append(alive, ex)
		}
	}
	g.explosions = alive
}

func (g *game) spawnExplosion(x, y, radius float64) {
	g.explosions = append(g.explosions, &explosion{x: x, y: y, radius: radius, life: 16, max: 16})
}

func (g *game) applyDefeatEnergyState() {
	if g.fort.hp <= 0 {
		g.fort.hp = 0
	}
	if g.player.hp <= 0 {
		g.player.hp = 0
		g.player.turretHP = 0
	}
}
