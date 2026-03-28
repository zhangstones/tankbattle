package game

func newPlayingGameForTest() *game {
	g := newGame()
	g.startMatch()
	return g
}
