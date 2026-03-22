package tankbattle

import "github.com/hajimehoshi/ebiten/v2"

func Run() error {
	ebiten.SetWindowSize(screenW, screenH)
	ebiten.SetWindowTitle("Tank Battle: Fortress Frontline")
	setWindowIcon()
	return ebiten.RunGame(newGame())
}
