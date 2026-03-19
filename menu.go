package tankbattle

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type menuAction int

const (
	menuNavUp menuAction = iota
	menuNavDown
	menuDec
	menuInc
	menuStart
	menuSetEasy
	menuSetNormal
	menuSetHard
	menuEnemyDown
	menuEnemyUp
)

func (g *game) updateMenu() {
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		g.applyMenuAction(menuNavUp)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		g.applyMenuAction(menuNavDown)
	}
	if inpututil.IsKeyJustPressed(ebiten.Key1) {
		g.applyMenuAction(menuSetEasy)
	}
	if inpututil.IsKeyJustPressed(ebiten.Key2) {
		g.applyMenuAction(menuSetNormal)
	}
	if inpututil.IsKeyJustPressed(ebiten.Key3) {
		g.applyMenuAction(menuSetHard)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyMinus) || inpututil.IsKeyJustPressed(ebiten.KeyNumpadSubtract) {
		g.applyMenuAction(menuEnemyDown)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEqual) || inpututil.IsKeyJustPressed(ebiten.KeyNumpadAdd) {
		g.applyMenuAction(menuEnemyUp)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.applyMenuAction(menuStart)
	}

	if g.menuIndex == 0 || g.menuIndex == 1 {
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) || inpututil.IsKeyJustPressed(ebiten.KeyA) {
			g.applyMenuAction(menuDec)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) || inpututil.IsKeyJustPressed(ebiten.KeyD) {
			g.applyMenuAction(menuInc)
		}
	}
}

func (g *game) applyMenuAction(action menuAction) {
	switch action {
	case menuNavUp:
		g.menuIndex--
		if g.menuIndex < 0 {
			g.menuIndex = menuItemCount - 1
		}
	case menuNavDown:
		g.menuIndex++
		if g.menuIndex >= menuItemCount {
			g.menuIndex = 0
		}
	case menuDec:
		if g.menuIndex == 0 && g.difficulty > diffEasy {
			g.difficulty--
		}
		if g.menuIndex == 1 && g.enemyBase > enemyBaseMin {
			g.enemyBase--
		}
	case menuInc:
		if g.menuIndex == 0 && g.difficulty < diffHard {
			g.difficulty++
		}
		if g.menuIndex == 1 && g.enemyBase < enemyBaseMax {
			g.enemyBase++
		}
	case menuSetEasy:
		g.difficulty = diffEasy
	case menuSetNormal:
		g.difficulty = diffNormal
	case menuSetHard:
		g.difficulty = diffHard
	case menuEnemyDown:
		if g.enemyBase > enemyBaseMin {
			g.enemyBase--
		}
	case menuEnemyUp:
		if g.enemyBase < enemyBaseMax {
			g.enemyBase++
		}
	case menuStart:
		g.startMatch()
	}
}
