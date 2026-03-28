package audio

import (
	"bytes"
	"io"
	"sync"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

const audioSampleRate = 44100

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

type Manager struct {
	ctx          *audio.Context
	enabled      bool
	masterVolume float64
	sfxVolume    float64
	maxPlayers   int
	lastPlayed   map[SFXID]int
	clips        map[SFXID]sfxClip
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

func NewManager() *Manager {
	a := &Manager{
		ctx:          sharedAudioContext(),
		enabled:      true,
		masterVolume: 0.9,
		sfxVolume:    1.0,
		maxPlayers:   16,
		lastPlayed:   map[SFXID]int{},
		clips:        map[SFXID]sfxClip{},
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

func (a *Manager) Enabled() bool {
	return a.enabled
}

func (a *Manager) SetEnabled(enabled bool) {
	a.enabled = enabled
}

func (a *Manager) SetSFXVolume(volume float64) {
	if volume < 0 {
		volume = 0
	}
	if volume > 1 {
		volume = 1
	}
	a.sfxVolume = volume
}

func (a *Manager) Play(id SFXID, frame int) {
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

func (a *Manager) sweepStoppedPlayers() {
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

func (a *Manager) enforceCapacity(group sfxGroup) {
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

func (a *Manager) duckMultiplier(group sfxGroup, frame int) float64 {
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

func (a *Manager) loadEmbeddedSFX() {
	type item struct {
		id       SFXID
		path     string
		cooldown int
		volume   float64
		group    sfxGroup
		priority int
	}
	manifest := []item{
		{id: SFXShootPlayer, path: "sfx/shoot_player.wav", cooldown: 5, volume: 0.68, group: sfxGroupCombat},
		{id: SFXShootEnemy, path: "sfx/shoot_enemy.wav", cooldown: 7, volume: 0.64, group: sfxGroupCombat},
		{id: SFXHitWall, path: "sfx/hit_wall.wav", cooldown: 3, volume: 0.56, group: sfxGroupCombat},
		{id: SFXHitTank, path: "sfx/hit_tank.wav", cooldown: 4, volume: 0.68, group: sfxGroupCombat},
		{id: SFXHitFortress, path: "sfx/hit_fortress.wav", cooldown: 5, volume: 0.74, group: sfxGroupCombat},
		{id: SFXExplosionSmall, path: "sfx/explosion_small.wav", cooldown: 5, volume: 0.66, group: sfxGroupCombat},
		{id: SFXExplosionLarge, path: "sfx/explosion_large.wav", cooldown: 9, volume: 0.84, group: sfxGroupCombat},
		{id: SFXPowerupSpawn, path: "sfx/powerup_spawn.wav", cooldown: 12, volume: 0.58, group: sfxGroupPowerup},
		{id: SFXPowerupPickupShield, path: "sfx/powerup_pickup_shield.wav", cooldown: 6, volume: 0.72, group: sfxGroupPowerup},
		{id: SFXPowerupPickupRapid, path: "sfx/powerup_pickup_rapid.wav", cooldown: 6, volume: 0.72, group: sfxGroupPowerup},
		{id: SFXPowerupPickupRepair, path: "sfx/powerup_pickup_repair.wav", cooldown: 6, volume: 0.72, group: sfxGroupPowerup},
		{id: SFXMenuMove, path: "sfx/menu_move.wav", cooldown: 2, volume: 0.50, group: sfxGroupUI},
		{id: SFXMenuConfirm, path: "sfx/menu_confirm.wav", cooldown: 3, volume: 0.62, group: sfxGroupUI},
		{id: SFXMenuBlocked, path: "sfx/menu_blocked.wav", cooldown: 4, volume: 0.56, group: sfxGroupUI},
		{id: SFXPauseToggle, path: "sfx/pause_toggle.wav", cooldown: 8, volume: 0.60, group: sfxGroupState, priority: 1},
		{id: SFXWavePrepare, path: "sfx/wave_prepare.wav", cooldown: 24, volume: 0.68, group: sfxGroupState, priority: 1},
		{id: SFXWaveStart, path: "sfx/wave_start.wav", cooldown: 24, volume: 0.74, group: sfxGroupState, priority: 1},
		{id: SFXBuffShieldOff, path: "sfx/buff_shield_off.wav", cooldown: 12, volume: 0.60, group: sfxGroupState, priority: 1},
		{id: SFXBuffRapidOff, path: "sfx/buff_rapid_off.wav", cooldown: 12, volume: 0.60, group: sfxGroupState, priority: 1},
		{id: SFXDestroyEnemy, path: "sfx/destroy_enemy.wav", cooldown: 10, volume: 0.80, group: sfxGroupCombat},
		{id: SFXDestroyPlayer, path: "sfx/destroy_player.wav", cooldown: 20, volume: 0.84, group: sfxGroupState, priority: 2},
		{id: SFXWin, path: "sfx/win.wav", cooldown: 40, volume: 0.86, group: sfxGroupState, priority: 2},
		{id: SFXLose, path: "sfx/lose.wav", cooldown: 40, volume: 0.86, group: sfxGroupState, priority: 2},
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
