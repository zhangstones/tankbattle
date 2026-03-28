package tankbattle

import (
	"os"

	gamepkg "tankbattle/internal/game"
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
	return gamepkg.RunWithOptions(gamepkg.RunOptions{
		DebugAPIAddr: opts.DebugAPIAddr,
	})
}
