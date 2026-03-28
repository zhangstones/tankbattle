package audio

import "testing"

func TestDuckMultiplierWindow(t *testing.T) {
	if duckMultiplier(sfxGroupCombat, 60) != 1.0 {
		t.Fatalf("combat duck should be disabled outside window")
	}
	if duckMultiplier(sfxGroupCombat, 10) >= 1.0 {
		t.Fatalf("combat duck should reduce gain during priority window")
	}
	if duckMultiplier(sfxGroupUI, 8) >= duckMultiplier(sfxGroupCombat, 8) {
		t.Fatalf("ui duck should be stronger than combat duck")
	}
	if duckMultiplier(sfxGroupState, 8) != 1.0 {
		t.Fatalf("state sounds should not be ducked")
	}
}

func TestAudioManagerLoadsWaveStateClips(t *testing.T) {
	a := NewManager()
	prepare, ok := a.clips[SFXWavePrepare]
	if !ok || len(prepare.pcm) == 0 {
		t.Fatalf("wave prepare clip should be loaded")
	}
	start, ok := a.clips[SFXWaveStart]
	if !ok || len(start.pcm) == 0 {
		t.Fatalf("wave start clip should be loaded")
	}
	if prepare.group != sfxGroupState || start.group != sfxGroupState {
		t.Fatalf("wave clips should belong to state group")
	}
}

func TestAudioManagerLoadsExtendedEventClips(t *testing.T) {
	a := NewManager()
	ids := []SFXID{
		SFXMenuBlocked,
		SFXBuffShieldOff,
		SFXBuffRapidOff,
		SFXDestroyEnemy,
		SFXDestroyPlayer,
		SFXPowerupPickupShield,
		SFXPowerupPickupRapid,
		SFXPowerupPickupRepair,
	}
	for _, id := range ids {
		clip, ok := a.clips[id]
		if !ok || len(clip.pcm) == 0 {
			t.Fatalf("expected clip to be loaded for id=%v", id)
		}
	}
}
