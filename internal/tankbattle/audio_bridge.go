package tankbattle

import gameaudio "tankbattle/internal/audio"

type sfxID = gameaudio.SFXID
type sfxPlayer = gameaudio.Player

const (
	sfxShootPlayer         = gameaudio.SFXShootPlayer
	sfxShootEnemy          = gameaudio.SFXShootEnemy
	sfxHitWall             = gameaudio.SFXHitWall
	sfxHitTank             = gameaudio.SFXHitTank
	sfxHitFortress         = gameaudio.SFXHitFortress
	sfxExplosionSmall      = gameaudio.SFXExplosionSmall
	sfxExplosionLarge      = gameaudio.SFXExplosionLarge
	sfxPowerupSpawn        = gameaudio.SFXPowerupSpawn
	sfxPowerupPickupShield = gameaudio.SFXPowerupPickupShield
	sfxPowerupPickupRapid  = gameaudio.SFXPowerupPickupRapid
	sfxPowerupPickupRepair = gameaudio.SFXPowerupPickupRepair
	sfxMenuMove            = gameaudio.SFXMenuMove
	sfxMenuConfirm         = gameaudio.SFXMenuConfirm
	sfxMenuBlocked         = gameaudio.SFXMenuBlocked
	sfxPauseToggle         = gameaudio.SFXPauseToggle
	sfxWin                 = gameaudio.SFXWin
	sfxLose                = gameaudio.SFXLose
	sfxWavePrepare         = gameaudio.SFXWavePrepare
	sfxWaveStart           = gameaudio.SFXWaveStart
	sfxBuffShieldOff       = gameaudio.SFXBuffShieldOff
	sfxBuffRapidOff        = gameaudio.SFXBuffRapidOff
	sfxDestroyEnemy        = gameaudio.SFXDestroyEnemy
	sfxDestroyPlayer       = gameaudio.SFXDestroyPlayer
)

func newAudioManager() sfxPlayer {
	return gameaudio.NewManager()
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
