package tankbattle

import (
	"bytes"
	"io"
	"sync"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

const audioSampleRate = 44100

type sfxID int

const (
	sfxShootPlayer sfxID = iota
	sfxShootEnemy
	sfxHitWall
	sfxHitTank
	sfxHitFortress
	sfxExplosionSmall
	sfxExplosionLarge
	sfxPowerupSpawn
	sfxPowerupPickup
	sfxMenuMove
	sfxMenuConfirm
	sfxPauseToggle
	sfxWin
	sfxLose
)

type sfxClip struct {
	pcm      []byte
	cooldown int
	volume   float64
}

type audioManager struct {
	ctx          *audio.Context
	enabled      bool
	masterVolume float64
	sfxVolume    float64
	maxPlayers   int
	lastPlayed   map[sfxID]int
	clips        map[sfxID]sfxClip
	players      []*audio.Player
}

var (
	audioContextOnce sync.Once
	sharedAudioCtx   *audio.Context
)

func newAudioManager() *audioManager {
	a := &audioManager{
		ctx:          sharedAudioContext(),
		enabled:      true,
		masterVolume: 0.9,
		sfxVolume:    1.0,
		maxPlayers:   16,
		lastPlayed:   map[sfxID]int{},
		clips:        map[sfxID]sfxClip{},
	}
	a.loadEmbeddedSFX()
	return a
}

func sharedAudioContext() *audio.Context {
	audioContextOnce.Do(func() {
		sharedAudioCtx = audio.NewContext(audioSampleRate)
	})
	return sharedAudioCtx
}

func (a *audioManager) Enabled() bool {
	return a.enabled
}

func (a *audioManager) SetEnabled(enabled bool) {
	a.enabled = enabled
}

func (a *audioManager) Play(id sfxID, frame int) {
	if a == nil || !a.enabled || a.ctx == nil {
		return
	}
	clip, ok := a.clips[id]
	if !ok || len(clip.pcm) == 0 {
		return
	}
	if last, hit := a.lastPlayed[id]; hit && frame-last < clip.cooldown {
		return
	}
	a.sweepStoppedPlayers()
	if len(a.players) >= a.maxPlayers {
		_ = a.players[0].Close()
		a.players = a.players[1:]
	}

	p := audio.NewPlayerFromBytes(a.ctx, clip.pcm)
	p.SetVolume(clip.volume * a.masterVolume * a.sfxVolume)
	p.Play()
	a.players = append(a.players, p)
	a.lastPlayed[id] = frame
}

func (a *audioManager) sweepStoppedPlayers() {
	alive := a.players[:0]
	for _, p := range a.players {
		if p.IsPlaying() {
			alive = append(alive, p)
			continue
		}
		_ = p.Close()
	}
	a.players = alive
}

func (a *audioManager) loadEmbeddedSFX() {
	type item struct {
		id       sfxID
		path     string
		cooldown int
		volume   float64
	}
	manifest := []item{
		{id: sfxShootPlayer, path: "assets/sfx/shoot_player.wav", cooldown: 5, volume: 0.70},
		{id: sfxShootEnemy, path: "assets/sfx/shoot_enemy.wav", cooldown: 7, volume: 0.66},
		{id: sfxHitWall, path: "assets/sfx/hit_wall.wav", cooldown: 3, volume: 0.58},
		{id: sfxHitTank, path: "assets/sfx/hit_tank.wav", cooldown: 4, volume: 0.65},
		{id: sfxHitFortress, path: "assets/sfx/hit_fortress.wav", cooldown: 5, volume: 0.72},
		{id: sfxExplosionSmall, path: "assets/sfx/explosion_small.wav", cooldown: 5, volume: 0.68},
		{id: sfxExplosionLarge, path: "assets/sfx/explosion_large.wav", cooldown: 8, volume: 0.78},
		{id: sfxPowerupSpawn, path: "assets/sfx/powerup_spawn.wav", cooldown: 10, volume: 0.52},
		{id: sfxPowerupPickup, path: "assets/sfx/powerup_pickup.wav", cooldown: 5, volume: 0.70},
		{id: sfxMenuMove, path: "assets/sfx/menu_move.wav", cooldown: 2, volume: 0.44},
		{id: sfxMenuConfirm, path: "assets/sfx/menu_confirm.wav", cooldown: 3, volume: 0.56},
		{id: sfxPauseToggle, path: "assets/sfx/pause_toggle.wav", cooldown: 6, volume: 0.55},
		{id: sfxWin, path: "assets/sfx/win.wav", cooldown: 30, volume: 0.80},
		{id: sfxLose, path: "assets/sfx/lose.wav", cooldown: 30, volume: 0.80},
	}
	for _, m := range manifest {
		raw, err := sfxFS.ReadFile(m.path)
		if err != nil {
			continue
		}
		decoded, err := wav.DecodeWithSampleRate(audioSampleRate, bytes.NewReader(raw))
		if err != nil {
			continue
		}
		pcm, err := io.ReadAll(decoded)
		if err != nil {
			continue
		}
		a.clips[m.id] = sfxClip{
			pcm:      pcm,
			cooldown: m.cooldown,
			volume:   m.volume,
		}
	}
}

func (g *game) playSFX(id sfxID) {
	if g == nil || g.audio == nil {
		return
	}
	g.audio.Play(id, g.frame)
}

func (g *game) setSoundEnabled(enabled bool) {
	g.soundEnabled = enabled
	if g.audio != nil {
		g.audio.SetEnabled(enabled)
	}
}

func (g *game) toggleSoundEnabled() {
	g.setSoundEnabled(!g.soundEnabled)
}
