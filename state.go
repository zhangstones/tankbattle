package tankbattle

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func (g *game) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyR) && g.restartIfAllowed() {
		return nil
	}

	switch g.state {
	case stateMenu:
		g.updateMenu()
		return nil
	case stateEnded:
		if inpututil.IsKeyJustPressed(ebiten.KeyM) {
			g.returnToMenu()
		}
		return nil
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyM) {
		g.returnToMenu()
		return nil
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
	}
	if g.rapidTick > 0 {
		g.rapidTick--
	}

	g.updatePlayer()
	g.updateEnemies()
	g.updateBullets()
	g.updatePowerups()
	g.updateExplosions()
	g.cleanupWalls()
	g.trySpawnRandomPowerup()

	if g.fort.hp <= 0 || g.player.hp <= 0 {
		g.state = stateEnded
		g.win = false
		return nil
	}

	if len(g.enemies) == 0 {
		if g.waveDelay == 0 {
			if g.wave >= g.maxWave {
				g.state = stateEnded
				g.win = true
			} else {
				g.wave++
				g.waveDelay = 130
				g.setMessage(fmt.Sprintf("Prepare wave %d", g.wave), g.waveDelay)
			}
		} else {
			g.waveDelay--
			if g.waveDelay == 0 {
				g.spawnWave(g.wave)
				g.setMessage(fmt.Sprintf("Wave %d incoming", g.wave), 100)
			}
		}
	}
	return nil
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
}

func (g *game) togglePause() {
	g.paused = !g.paused
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
