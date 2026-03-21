package tankbattle

type mockSFXPlayer struct {
	enabled bool
	played  []sfxID
	frames  []int
}

func (m *mockSFXPlayer) Play(id sfxID, frame int) {
	m.played = append(m.played, id)
	m.frames = append(m.frames, frame)
}

func (m *mockSFXPlayer) SetEnabled(enabled bool) {
	m.enabled = enabled
}

func (m *mockSFXPlayer) SetSFXVolume(_ float64) {}

func (m *mockSFXPlayer) Enabled() bool {
	return m.enabled
}

func (m *mockSFXPlayer) last() (sfxID, bool) {
	if len(m.played) == 0 {
		return 0, false
	}
	return m.played[len(m.played)-1], true
}

func (m *mockSFXPlayer) lastFrame() (int, bool) {
	if len(m.frames) == 0 {
		return 0, false
	}
	return m.frames[len(m.frames)-1], true
}
