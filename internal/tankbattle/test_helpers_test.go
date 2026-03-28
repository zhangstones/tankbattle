package tankbattle

func newPlayingGameForTest() *game {
	g := newGame()
	g.startMatch()
	return g
}
