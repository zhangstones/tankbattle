package main

import (
	"encoding/binary"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

const sampleRate = 44100

type stereoSample struct {
	l float64
	r float64
}

type clipParams struct {
	name      string
	seconds   float64
	targetRMS float64
	peakLimit float64
	gen       func(t float64) stereoSample
}

func main() {
	rand.Seed(time.Now().UnixNano())
	outDir := filepath.Join("assets", "sfx")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		panic(err)
	}

	clips := []clipParams{
		{name: "shoot_player.wav", seconds: 0.22, targetRMS: 0.17, peakLimit: 0.88, gen: genShootPlayer},
		{name: "shoot_enemy.wav", seconds: 0.24, targetRMS: 0.17, peakLimit: 0.88, gen: genShootEnemy},
		{name: "hit_wall.wav", seconds: 0.18, targetRMS: 0.16, peakLimit: 0.86, gen: genHitWall},
		{name: "hit_tank.wav", seconds: 0.25, targetRMS: 0.17, peakLimit: 0.88, gen: genHitTank},
		{name: "hit_fortress.wav", seconds: 0.40, targetRMS: 0.19, peakLimit: 0.90, gen: genHitFortress},
		{name: "explosion_small.wav", seconds: 0.46, targetRMS: 0.19, peakLimit: 0.90, gen: genExplosionSmall},
		{name: "explosion_large.wav", seconds: 0.82, targetRMS: 0.20, peakLimit: 0.90, gen: genExplosionLarge},
		{name: "powerup_spawn.wav", seconds: 0.42, targetRMS: 0.16, peakLimit: 0.85, gen: genPowerupSpawn},
		{name: "powerup_pickup_shield.wav", seconds: 0.34, targetRMS: 0.16, peakLimit: 0.85, gen: genPowerupPickupShield},
		{name: "powerup_pickup_rapid.wav", seconds: 0.34, targetRMS: 0.16, peakLimit: 0.85, gen: genPowerupPickupRapid},
		{name: "powerup_pickup_repair.wav", seconds: 0.34, targetRMS: 0.16, peakLimit: 0.85, gen: genPowerupPickupRepair},
		{name: "menu_move.wav", seconds: 0.10, targetRMS: 0.11, peakLimit: 0.75, gen: genMenuMove},
		{name: "menu_confirm.wav", seconds: 0.18, targetRMS: 0.13, peakLimit: 0.80, gen: genMenuConfirm},
		{name: "menu_blocked.wav", seconds: 0.12, targetRMS: 0.12, peakLimit: 0.78, gen: genMenuBlocked},
		{name: "pause_toggle.wav", seconds: 0.16, targetRMS: 0.13, peakLimit: 0.80, gen: genPauseToggle},
		{name: "wave_prepare.wav", seconds: 0.33, targetRMS: 0.16, peakLimit: 0.84, gen: genWavePrepare},
		{name: "wave_start.wav", seconds: 0.44, targetRMS: 0.17, peakLimit: 0.86, gen: genWaveStart},
		{name: "buff_shield_off.wav", seconds: 0.22, targetRMS: 0.14, peakLimit: 0.82, gen: genBuffShieldOff},
		{name: "buff_rapid_off.wav", seconds: 0.22, targetRMS: 0.14, peakLimit: 0.82, gen: genBuffRapidOff},
		{name: "destroy_enemy.wav", seconds: 0.52, targetRMS: 0.19, peakLimit: 0.90, gen: genDestroyEnemy},
		{name: "destroy_player.wav", seconds: 0.66, targetRMS: 0.20, peakLimit: 0.90, gen: genDestroyPlayer},
		{name: "win.wav", seconds: 1.00, targetRMS: 0.18, peakLimit: 0.88, gen: genWin},
		{name: "lose.wav", seconds: 1.08, targetRMS: 0.18, peakLimit: 0.88, gen: genLose},
	}

	for _, c := range clips {
		samples := render(c.seconds, c.gen)
		normalize(&samples, c.targetRMS, c.peakLimit)
		writeWAV(filepath.Join(outDir, c.name), samples)
	}
}

func genShootPlayer(t float64) stereoSample {
	transient := filteredNoise(t, 2200, 7200) * expDecay(t, 0.03) * 0.72
	body := sineSweep(148, 92, t) * expDecay(t, 0.15) * 0.48
	mech := ring(890, t, 0.06, 0.22)
	return pan(transient+body+mech, 0.06)
}

func genShootEnemy(t float64) stereoSample {
	transient := filteredNoise(t, 1600, 5600) * expDecay(t, 0.04) * 0.66
	body := sineSweep(116, 72, t) * expDecay(t, 0.17) * 0.52
	mech := ring(620, t, 0.08, 0.20)
	return pan(transient+body+mech, -0.06)
}

func genHitWall(t float64) stereoSample {
	click := filteredNoise(t, 2500, 7600) * expDecay(t, 0.028) * 0.78
	grit := filteredNoise(t, 480, 1700) * expDecay(t, 0.11) * 0.38
	return pan(click+grit, jitter(t)*0.3)
}

func genHitTank(t float64) stereoSample {
	ping := ring(520, t, 0.05, 0.36)
	ping2 := ring(780, t, 0.07, 0.26)
	impact := filteredNoise(t, 1500, 4600) * expDecay(t, 0.06) * 0.52
	return pan(ping+ping2+impact, jitter(t)*0.2)
}

func genHitFortress(t float64) stereoSample {
	punch := filteredNoise(t, 900, 3200) * expDecay(t, 0.06) * 0.58
	body := sineSweep(96, 54, t) * expDecay(t, 0.24) * 0.54
	tail := filteredNoise(t, 100, 500) * expDecay(t, 0.30) * 0.34
	return pan(punch+body+tail, 0)
}

func genExplosionSmall(t float64) stereoSample {
	blast := filteredNoise(t, 120, 5200) * expDecay(t, 0.14) * 0.86
	low := sineSweep(88, 44, t) * expDecay(t, 0.25) * 0.46
	air := filteredNoise(t, 800, 3000) * expDecay(t, 0.22) * 0.22
	return pan(blast+low+air, jitter(t)*0.12)
}

func genExplosionLarge(t float64) stereoSample {
	blast := filteredNoise(t, 80, 5800) * expDecay(t, 0.22) * 0.88
	low := sineSweep(70, 30, t) * expDecay(t, 0.46) * 0.68
	tail := filteredNoise(t, 70, 1200) * expDecay(t, 0.68) * 0.36
	return pan(blast+low+tail, jitter(t)*0.10)
}

func genPowerupSpawn(t float64) stereoSample {
	env := attackExp(t, 0.02, 0.30)
	chime := math.Sin(2*math.Pi*(380+520*t)*t) * env * 0.36
	chime2 := math.Sin(2*math.Pi*(760+300*t)*t) * expDecay(t, 0.24) * 0.20
	air := filteredNoise(t, 2400, 5200) * expDecay(t, 0.20) * 0.20
	return pan(chime+chime2+air, math.Sin(t*7.5)*0.12)
}

func genPowerupPickupShield(t float64) stereoSample {
	env := attackExp(t, 0.01, 0.22)
	n1 := math.Sin(2*math.Pi*620*t) * env * 0.33
	n2 := math.Sin(2*math.Pi*930*t) * expDecay(t, 0.18) * 0.23
	n3 := math.Sin(2*math.Pi*1240*t) * expDecay(t, 0.14) * 0.14
	return pan(n1+n2+n3, -math.Sin(t*8)*0.08)
}

func genPowerupPickupRapid(t float64) stereoSample {
	env := attackExp(t, 0.008, 0.20)
	n1 := math.Sin(2*math.Pi*720*t) * env * 0.34
	n2 := math.Sin(2*math.Pi*1080*t) * expDecay(t, 0.16) * 0.24
	n3 := math.Sin(2*math.Pi*1440*t) * expDecay(t, 0.13) * 0.15
	return pan(n1+n2+n3, -math.Sin(t*8)*0.08)
}

func genPowerupPickupRepair(t float64) stereoSample {
	env := attackExp(t, 0.012, 0.24)
	n1 := math.Sin(2*math.Pi*560*t) * env * 0.34
	n2 := math.Sin(2*math.Pi*840*t) * expDecay(t, 0.19) * 0.24
	n3 := math.Sin(2*math.Pi*1120*t) * expDecay(t, 0.15) * 0.15
	return pan(n1+n2+n3, -math.Sin(t*8)*0.08)
}

func genMenuMove(t float64) stereoSample {
	tick := ring(1260, t, 0.05, 0.24) + filteredNoise(t, 2800, 6000)*expDecay(t, 0.03)*0.18
	return pan(tick, 0)
}

func genMenuConfirm(t float64) stereoSample {
	click := filteredNoise(t, 1700, 5200) * expDecay(t, 0.04) * 0.30
	body := ring(520, t, 0.10, 0.30)
	return pan(click+body, 0)
}

func genMenuBlocked(t float64) stereoSample {
	fall := sineSweep(460, 290, t) * expDecay(t, 0.10) * 0.28
	grit := filteredNoise(t, 1400, 3400) * expDecay(t, 0.05) * 0.20
	return pan(fall+grit, 0)
}

func genPauseToggle(t float64) stereoSample {
	pulse := sineSweep(760, 430, t) * expDecay(t, 0.12) * 0.33
	return pan(pulse, 0)
}

func genWavePrepare(t float64) stereoSample {
	base := math.Sin(2*math.Pi*420*t) * attackExp(t, 0.01, 0.26) * 0.26
	up := math.Sin(2*math.Pi*(420+260*t)*t) * expDecay(t, 0.24) * 0.20
	return pan(base+up, 0)
}

func genWaveStart(t float64) stereoSample {
	ping := math.Sin(2*math.Pi*540*t) * attackExp(t, 0.01, 0.28) * 0.26
	ping2 := math.Sin(2*math.Pi*810*t) * expDecay(t, 0.22) * 0.18
	body := filteredNoise(t, 700, 2200) * expDecay(t, 0.10) * 0.22
	return pan(ping+ping2+body, 0)
}

func genBuffShieldOff(t float64) stereoSample {
	fall := sineSweep(640, 340, t) * expDecay(t, 0.16) * 0.26
	shimmer := filteredNoise(t, 2200, 4800) * expDecay(t, 0.08) * 0.14
	return pan(fall+shimmer, 0)
}

func genBuffRapidOff(t float64) stereoSample {
	fall := sineSweep(760, 420, t) * expDecay(t, 0.14) * 0.25
	shimmer := filteredNoise(t, 2000, 4200) * expDecay(t, 0.07) * 0.14
	return pan(fall+shimmer, 0)
}

func genDestroyEnemy(t float64) stereoSample {
	blast := filteredNoise(t, 120, 4600) * expDecay(t, 0.16) * 0.86
	low := sineSweep(98, 42, t) * expDecay(t, 0.28) * 0.52
	return pan(blast+low, jitter(t)*0.12)
}

func genDestroyPlayer(t float64) stereoSample {
	blast := filteredNoise(t, 90, 5200) * expDecay(t, 0.22) * 0.86
	low := sineSweep(82, 28, t) * expDecay(t, 0.40) * 0.62
	tail := filteredNoise(t, 70, 1000) * expDecay(t, 0.48) * 0.26
	return pan(blast+low+tail, 0)
}

func genWin(t float64) stereoSample {
	scale := []float64{523.25, 659.25, 783.99, 1046.5}
	slot := int(t / 0.24)
	if slot >= len(scale) {
		slot = len(scale) - 1
	}
	local := math.Mod(t, 0.24)
	n := math.Sin(2*math.Pi*scale[slot]*t) * attackExp(local, 0.01, 0.20) * 0.35
	h := math.Sin(2*math.Pi*scale[slot]*2*t) * expDecay(local, 0.16) * 0.12
	return pan(n+h, math.Sin(t*2.2)*0.16)
}

func genLose(t float64) stereoSample {
	base := 240 - 130*t
	main := math.Sin(2*math.Pi*base*t) * expDecay(t, 0.84) * 0.36
	low := math.Sin(2*math.Pi*(base*0.52)*t) * expDecay(t, 0.92) * 0.22
	noise := filteredNoise(t, 160, 980) * expDecay(t, 0.52) * 0.20
	return pan(main+low+noise, 0)
}

func render(seconds float64, fn func(t float64) stereoSample) []stereoSample {
	n := int(seconds * sampleRate)
	out := make([]stereoSample, n)
	for i := 0; i < n; i++ {
		t := float64(i) / sampleRate
		out[i] = fn(t)
	}
	return out
}

func normalize(samples *[]stereoSample, targetRMS, peakLimit float64) {
	if len(*samples) == 0 {
		return
	}
	sum := 0.0
	peak := 0.0
	for _, s := range *samples {
		m := (s.l + s.r) * 0.5
		sum += m * m
		peak = maxF(peak, math.Abs(s.l))
		peak = maxF(peak, math.Abs(s.r))
	}
	rms := math.Sqrt(sum / float64(len(*samples)))
	if rms < 1e-6 || peak < 1e-6 {
		return
	}

	rmsGain := targetRMS / rms
	peakGain := peakLimit / peak
	gain := math.Min(rmsGain, peakGain)
	for i := range *samples {
		(*samples)[i].l = clamp((*samples)[i].l * gain)
		(*samples)[i].r = clamp((*samples)[i].r * gain)
	}
}

func ring(freq, t, tau, amp float64) float64 {
	return math.Sin(2*math.Pi*freq*t) * expDecay(t, tau) * amp
}

func sineSweep(startHz, endHz, t float64) float64 {
	f := startHz + (endHz-startHz)*t*1.8
	return math.Sin(2 * math.Pi * f * t)
}

func filteredNoise(t, lowHz, highHz float64) float64 {
	mid := (lowHz + highHz) * 0.5
	shape := math.Sin(2*math.Pi*lowHz*t+math.Sin(t*13.4)*0.6)*0.34 +
		math.Sin(2*math.Pi*mid*t+math.Sin(t*7.8)*0.5)*0.42 +
		math.Sin(2*math.Pi*highHz*t+math.Sin(t*5.3)*0.4)*0.24
	shape += (rand.Float64()*2 - 1) * 0.08
	return shape
}

func attackExp(t, attack, tau float64) float64 {
	if t < 0 {
		return 0
	}
	a := t / attack
	if a > 1 {
		a = 1
	}
	return a * expDecay(t, tau)
}

func expDecay(t, tau float64) float64 {
	if t < 0 {
		return 0
	}
	return math.Exp(-t / tau)
}

func pan(v, p float64) stereoSample {
	if p > 1 {
		p = 1
	}
	if p < -1 {
		p = -1
	}
	l := v * (1 - p*0.32)
	r := v * (1 + p*0.32)
	return stereoSample{l: l, r: r}
}

func jitter(t float64) float64 {
	return math.Sin(t*13.2)*0.5 + math.Sin(t*8.1)*0.3
}

func clamp(v float64) float64 {
	if v > 1 {
		return 1
	}
	if v < -1 {
		return -1
	}
	return v
}

func writeWAV(path string, samples []stereoSample) {
	const (
		channels      = 2
		bitsPerSample = 16
	)
	blockAlign := channels * bitsPerSample / 8
	byteRate := sampleRate * blockAlign
	dataSize := len(samples) * blockAlign
	riffSize := 36 + dataSize

	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	mustWrite(f, []byte("RIFF"))
	mustWriteU32(f, uint32(riffSize))
	mustWrite(f, []byte("WAVE"))
	mustWrite(f, []byte("fmt "))
	mustWriteU32(f, 16)
	mustWriteU16(f, 1)
	mustWriteU16(f, channels)
	mustWriteU32(f, sampleRate)
	mustWriteU32(f, uint32(byteRate))
	mustWriteU16(f, uint16(blockAlign))
	mustWriteU16(f, bitsPerSample)
	mustWrite(f, []byte("data"))
	mustWriteU32(f, uint32(dataSize))

	for _, s := range samples {
		l := int16(clamp(s.l) * 32767)
		r := int16(clamp(s.r) * 32767)
		mustWriteI16(f, l)
		mustWriteI16(f, r)
	}
}

func mustWrite(f *os.File, b []byte) {
	if _, err := f.Write(b); err != nil {
		panic(err)
	}
}

func mustWriteU16(f *os.File, v uint16) {
	var b [2]byte
	binary.LittleEndian.PutUint16(b[:], v)
	mustWrite(f, b[:])
}

func mustWriteU32(f *os.File, v uint32) {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], v)
	mustWrite(f, b[:])
}

func mustWriteI16(f *os.File, v int16) {
	var b [2]byte
	binary.LittleEndian.PutUint16(b[:], uint16(v))
	mustWrite(f, b[:])
}

func maxF(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
