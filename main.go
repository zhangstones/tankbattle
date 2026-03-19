package tankbattle

import "github.com/hajimehoshi/ebiten/v2"

func Run() error {
	ebiten.SetWindowSize(screenW, screenH)
	ebiten.SetWindowTitle("Tank Battle: Fortress Frontline")
	return ebiten.RunGame(newGame())
}
