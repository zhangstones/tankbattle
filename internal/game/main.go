package game

import (
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

type RunOptions struct {
	DebugAPIAddr string
}

func Run() error {
	return RunWithOptions(RunOptions{
		DebugAPIAddr: os.Getenv("TANKBATTLE_DEBUG_API_ADDR"),
	})
}

func RunWithOptions(opts RunOptions) error {
	var debug *DebugController
	gameOptions := newGameOptions{
		loadUserSettings: true,
		persistUserData:  true,
		audio:            newAudioManager(),
	}
	if opts.DebugAPIAddr != "" {
		debug = NewDebugController()
		gameOptions.loadUserSettings = false
		gameOptions.persistUserData = false
		gameOptions.audio = nil
		gameOptions.debug = debug
		seed := int64(20260328)
		gameOptions.randomSeed = &seed
	}
	g := newGameWithOptions(gameOptions)
	if debug != nil {
		if err := debug.StartHTTP(opts.DebugAPIAddr); err != nil {
			return err
		}
		defer debug.Close()
	}

	ebiten.SetWindowSize(screenW, screenH)
	ebiten.SetWindowTitle("Tank Battle: Fortress Frontline")
	ebiten.SetRunnableOnUnfocused(debug != nil)
	setWindowIcon()
	return ebiten.RunGame(g)
}
