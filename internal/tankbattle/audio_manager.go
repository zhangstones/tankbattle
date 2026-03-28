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
	sfxPowerupPickupShield
	sfxPowerupPickupRapid
	sfxPowerupPickupRepair
	sfxMenuMove
	sfxMenuConfirm
	sfxMenuBlocked
	sfxPauseToggle
	sfxWin
	sfxLose
	sfxWavePrepare
	sfxWaveStart
	sfxBuffShieldOff
	sfxBuffRapidOff
	sfxDestroyEnemy
	sfxDestroyPlayer
)

type sfxGroup int

const (
	sfxGroupCombat sfxGroup = iota
	sfxGroupPowerup
	sfxGroupUI
	sfxGroupState
)

type sfxClip struct {
	pcm      []byte
	cooldown int
	volume   float64
	group    sfxGroup
	priority int
}

type audioManager struct {
	ctx          *audio.Context
	enabled      bool
	masterVolume float64
	sfxVolume    float64
	maxPlayers   int
	lastPlayed   map[sfxID]int
	clips        map[sfxID]sfxClip
	players      []activePlayer
	groupGain    map[sfxGroup]float64
	groupCap     map[sfxGroup]int
	priorityTick int
}

type activePlayer struct {
	group  sfxGroup
	player *audio.Player
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
		groupGain: map[sfxGroup]float64{
			sfxGroupCombat:  1.0,
			sfxGroupPowerup: 0.86,
			sfxGroupUI:      0.62,
			sfxGroupState:   0.82,
		},
		groupCap: map[sfxGroup]int{
			sfxGroupCombat:  10,
			sfxGroupPowerup: 3,
			sfxGroupUI:      3,
			sfxGroupState:   2,
		},
		priorityTick: -9999,
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

func (a *audioManager) SetSFXVolume(volume float64) {
	if volume < 0 {
		volume = 0
	}
	if volume > 1 {
		volume = 1
	}
	a.sfxVolume = volume
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
	a.enforceCapacity(clip.group)

	p := audio.NewPlayerFromBytes(a.ctx, clip.pcm)
	duck := a.duckMultiplier(clip.group, frame)
	groupGain := a.groupGain[clip.group]
	p.SetVolume(clip.volume * groupGain * duck * a.masterVolume * a.sfxVolume)
	p.Play()
	a.players = append(a.players, activePlayer{group: clip.group, player: p})
	a.lastPlayed[id] = frame
	if clip.priority > 0 {
		a.priorityTick = frame
	}
}

func (a *audioManager) sweepStoppedPlayers() {
	alive := a.players[:0]
	for _, entry := range a.players {
		if entry.player.IsPlaying() {
			alive = append(alive, entry)
			continue
		}
		_ = entry.player.Close()
	}
	a.players = alive
}

func (a *audioManager) enforceCapacity(group sfxGroup) {
	if len(a.players) >= a.maxPlayers {
		_ = a.players[0].player.Close()
		a.players = a.players[1:]
	}
	capByGroup := a.groupCap[group]
	if capByGroup <= 0 {
		return
	}
	count := 0
	for _, entry := range a.players {
		if entry.group == group {
			count++
		}
	}
	if count < capByGroup {
		return
	}
	for i, entry := range a.players {
		if entry.group != group {
			continue
		}
		_ = entry.player.Close()
		a.players = append(a.players[:i], a.players[i+1:]...)
		return
	}
}

func (a *audioManager) duckMultiplier(group sfxGroup, frame int) float64 {
	if group == sfxGroupState {
		return 1.0
	}
	since := frame - a.priorityTick
	return duckMultiplier(group, since)
}

func duckMultiplier(group sfxGroup, framesSincePriority int) float64 {
	if framesSincePriority < 0 || framesSincePriority > 45 {
		return 1.0
	}
	if group == sfxGroupUI {
		return 0.60
	}
	if group == sfxGroupCombat || group == sfxGroupPowerup {
		return 0.76
	}
	return 1.0
}

func (a *audioManager) loadEmbeddedSFX() {
	type item struct {
		id       sfxID
		path     string
		cooldown int
		volume   float64
		group    sfxGroup
		priority int
	}
	manifest := []item{
		{id: sfxShootPlayer, path: "sfx/shoot_player.wav", cooldown: 5, volume: 0.68, group: sfxGroupCombat},
		{id: sfxShootEnemy, path: "sfx/shoot_enemy.wav", cooldown: 7, volume: 0.64, group: sfxGroupCombat},
		{id: sfxHitWall, path: "sfx/hit_wall.wav", cooldown: 3, volume: 0.56, group: sfxGroupCombat},
		{id: sfxHitTank, path: "sfx/hit_tank.wav", cooldown: 4, volume: 0.68, group: sfxGroupCombat},
		{id: sfxHitFortress, path: "sfx/hit_fortress.wav", cooldown: 5, volume: 0.74, group: sfxGroupCombat},
		{id: sfxExplosionSmall, path: "sfx/explosion_small.wav", cooldown: 5, volume: 0.66, group: sfxGroupCombat},
		{id: sfxExplosionLarge, path: "sfx/explosion_large.wav", cooldown: 9, volume: 0.84, group: sfxGroupCombat},
		{id: sfxPowerupSpawn, path: "sfx/powerup_spawn.wav", cooldown: 12, volume: 0.58, group: sfxGroupPowerup},
		{id: sfxPowerupPickupShield, path: "sfx/powerup_pickup_shield.wav", cooldown: 6, volume: 0.72, group: sfxGroupPowerup},
		{id: sfxPowerupPickupRapid, path: "sfx/powerup_pickup_rapid.wav", cooldown: 6, volume: 0.72, group: sfxGroupPowerup},
		{id: sfxPowerupPickupRepair, path: "sfx/powerup_pickup_repair.wav", cooldown: 6, volume: 0.72, group: sfxGroupPowerup},
		{id: sfxMenuMove, path: "sfx/menu_move.wav", cooldown: 2, volume: 0.50, group: sfxGroupUI},
		{id: sfxMenuConfirm, path: "sfx/menu_confirm.wav", cooldown: 3, volume: 0.62, group: sfxGroupUI},
		{id: sfxMenuBlocked, path: "sfx/menu_blocked.wav", cooldown: 4, volume: 0.56, group: sfxGroupUI},
		{id: sfxPauseToggle, path: "sfx/pause_toggle.wav", cooldown: 8, volume: 0.60, group: sfxGroupState, priority: 1},
		{id: sfxWavePrepare, path: "sfx/wave_prepare.wav", cooldown: 24, volume: 0.68, group: sfxGroupState, priority: 1},
		{id: sfxWaveStart, path: "sfx/wave_start.wav", cooldown: 24, volume: 0.74, group: sfxGroupState, priority: 1},
		{id: sfxBuffShieldOff, path: "sfx/buff_shield_off.wav", cooldown: 12, volume: 0.60, group: sfxGroupState, priority: 1},
		{id: sfxBuffRapidOff, path: "sfx/buff_rapid_off.wav", cooldown: 12, volume: 0.60, group: sfxGroupState, priority: 1},
		{id: sfxDestroyEnemy, path: "sfx/destroy_enemy.wav", cooldown: 10, volume: 0.80, group: sfxGroupCombat},
		{id: sfxDestroyPlayer, path: "sfx/destroy_player.wav", cooldown: 20, volume: 0.84, group: sfxGroupState, priority: 2},
		{id: sfxWin, path: "sfx/win.wav", cooldown: 40, volume: 0.86, group: sfxGroupState, priority: 2},
		{id: sfxLose, path: "sfx/lose.wav", cooldown: 40, volume: 0.86, group: sfxGroupState, priority: 2},
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
			group:    m.group,
			priority: m.priority,
		}
	}
}

func (g *game) playSFX(id sfxID) {
	if g == nil || g.audio == nil {
		return
	}
	g.audio.Play(id, g.audioFrame)
}

func (g *game) setSoundEnabled(enabled bool) {
	g.soundEnabled = enabled
	if g.audio != nil {
		g.audio.SetEnabled(enabled)
	}
	g.saveUserSettings()
}

func (g *game) toggleSoundEnabled() {
	g.setSoundEnabled(!g.soundEnabled)
}

func (g *game) setSoundVolumePercent(percent int) {
	percent = clampInt(percent, 0, 100)
	g.soundVolume = percent
	if g.audio != nil {
		g.audio.SetSFXVolume(float64(percent) / 100.0)
	}
	g.saveUserSettings()
}

func (g *game) adjustSoundVolume(delta int) bool {
	next := clampInt(g.soundVolume+delta, 0, 100)
	if next == g.soundVolume {
		return false
	}
	g.setSoundVolumePercent(next)
	return true
}
