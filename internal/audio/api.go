package audio

type SFXID int

const (
	SFXShootPlayer SFXID = iota
	SFXShootEnemy
	SFXHitWall
	SFXHitTank
	SFXHitFortress
	SFXExplosionSmall
	SFXExplosionLarge
	SFXPowerupSpawn
	SFXPowerupPickupShield
	SFXPowerupPickupRapid
	SFXPowerupPickupRepair
	SFXMenuMove
	SFXMenuConfirm
	SFXMenuBlocked
	SFXPauseToggle
	SFXWin
	SFXLose
	SFXWavePrepare
	SFXWaveStart
	SFXBuffShieldOff
	SFXBuffRapidOff
	SFXDestroyEnemy
	SFXDestroyPlayer
)

type Player interface {
	Play(id SFXID, frame int)
	SetEnabled(enabled bool)
	SetSFXVolume(volume float64)
	Enabled() bool
}
