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

func main() {
	rand.Seed(time.Now().UnixNano())
	outDir := filepath.Join("assets", "sfx")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		panic(err)
	}

	writeWAV(filepath.Join(outDir, "shoot_player.wav"), render(0.20, func(t float64) stereoSample {
		env := decay(t, 0.11)
		crack := bandNoise(t, 1600, 3800, 0.85) * env
		thump := math.Sin(2*math.Pi*120*t) * decay(t, 0.18) * 0.30
		recoil := math.Sin(2*math.Pi*(760-220*t)*t) * decay(t, 0.09) * 0.36
		return pan(crack+thump+recoil, 0.02)
	}))

	writeWAV(filepath.Join(outDir, "shoot_enemy.wav"), render(0.22, func(t float64) stereoSample {
		env := decay(t, 0.14)
		crack := bandNoise(t, 900, 2600, 0.8) * env
		thump := math.Sin(2*math.Pi*86*t) * decay(t, 0.2) * 0.40
		recoil := math.Sin(2*math.Pi*(520-150*t)*t) * decay(t, 0.12) * 0.28
		return pan(crack+thump+recoil, -0.03)
	}))

	writeWAV(filepath.Join(outDir, "hit_wall.wav"), render(0.17, func(t float64) stereoSample {
		env := decay(t, 0.07)
		click := bandNoise(t, 2000, 5200, 0.95) * env
		dust := bandNoise(t, 350, 1200, 0.55) * decay(t, 0.11)
		return pan(click+dust, jitterPan(t))
	}))

	writeWAV(filepath.Join(outDir, "hit_tank.wav"), render(0.24, func(t float64) stereoSample {
		env := decay(t, 0.14)
		ringA := math.Sin(2*math.Pi*420*t) * env * 0.40
		ringB := math.Sin(2*math.Pi*680*t) * decay(t, 0.10) * 0.24
		shard := bandNoise(t, 1400, 4000, 0.9) * decay(t, 0.08) * 0.44
		return pan(ringA+ringB+shard, jitterPan(t)*0.5)
	}))

	writeWAV(filepath.Join(outDir, "hit_fortress.wav"), render(0.38, func(t float64) stereoSample {
		boom := math.Sin(2*math.Pi*(74-12*t)*t) * decay(t, 0.24) * 0.56
		impact := bandNoise(t, 700, 2600, 0.9) * decay(t, 0.09) * 0.55
		rumble := bandNoise(t, 80, 260, 0.6) * decay(t, 0.30) * 0.38
		return pan(boom+impact+rumble, 0)
	}))

	writeWAV(filepath.Join(outDir, "explosion_small.wav"), render(0.42, func(t float64) stereoSample {
		blast := bandNoise(t, 110, 3600, 0.98) * decay(t, 0.17)
		low := math.Sin(2*math.Pi*(95-30*t)*t) * decay(t, 0.28) * 0.38
		return pan(blast+low, jitterPan(t)*0.3)
	}))

	writeWAV(filepath.Join(outDir, "explosion_large.wav"), render(0.72, func(t float64) stereoSample {
		blast := bandNoise(t, 80, 3200, 1.0) * decay(t, 0.29)
		low := math.Sin(2*math.Pi*(62-18*t)*t) * decay(t, 0.44) * 0.58
		tail := bandNoise(t, 60, 900, 0.75) * decay(t, 0.62) * 0.28
		return pan(blast+low+tail, jitterPan(t)*0.22)
	}))

	writeWAV(filepath.Join(outDir, "powerup_spawn.wav"), render(0.42, func(t float64) stereoSample {
		env := attackDecay(t, 0.06, 0.30)
		f0 := 340 + 620*t
		spark := math.Sin(2*math.Pi*f0*t) * env * 0.38
		spark2 := math.Sin(2*math.Pi*(f0*1.5)*t) * env * 0.2
		air := bandNoise(t, 2500, 5200, 0.4) * decay(t, 0.22) * 0.3
		return pan(spark+spark2+air, math.Sin(t*7)*0.15)
	}))

	writeWAV(filepath.Join(outDir, "powerup_pickup.wav"), render(0.32, func(t float64) stereoSample {
		env := attackDecay(t, 0.01, 0.22)
		n1 := math.Sin(2*math.Pi*660*t) * env * 0.42
		n2 := math.Sin(2*math.Pi*990*t) * decay(t, 0.18) * 0.26
		n3 := math.Sin(2*math.Pi*1320*t) * decay(t, 0.14) * 0.18
		return pan(n1+n2+n3, -math.Sin(t*8)*0.1)
	}))

	writeWAV(filepath.Join(outDir, "menu_move.wav"), render(0.11, func(t float64) stereoSample {
		env := decay(t, 0.06)
		tick := math.Sin(2*math.Pi*1240*t) * env * 0.24
		return pan(tick, 0)
	}))

	writeWAV(filepath.Join(outDir, "menu_confirm.wav"), render(0.19, func(t float64) stereoSample {
		env := decay(t, 0.12)
		body := math.Sin(2*math.Pi*520*t) * env * 0.28
		click := bandNoise(t, 1600, 4400, 0.5) * decay(t, 0.05) * 0.28
		return pan(body+click, 0)
	}))

	writeWAV(filepath.Join(outDir, "pause_toggle.wav"), render(0.16, func(t float64) stereoSample {
		env := decay(t, 0.1)
		pulse := math.Sin(2*math.Pi*(780-260*t)*t) * env * 0.26
		return pan(pulse, 0)
	}))

	writeWAV(filepath.Join(outDir, "win.wav"), render(0.92, func(t float64) stereoSample {
		scale := []float64{523.25, 659.25, 783.99, 1046.5}
		idx := int(t / 0.2)
		if idx >= len(scale) {
			idx = len(scale) - 1
		}
		note := math.Sin(2*math.Pi*scale[idx]*t) * attackDecay(math.Mod(t, 0.2), 0.01, 0.18) * 0.34
		harm := math.Sin(2*math.Pi*scale[idx]*2*t) * decay(math.Mod(t, 0.2), 0.14) * 0.12
		return pan(note+harm, math.Sin(t*2.2)*0.2)
	}))

	writeWAV(filepath.Join(outDir, "lose.wav"), render(0.95, func(t float64) stereoSample {
		base := 240 - 120*t
		env := decay(t, 0.75)
		main := math.Sin(2*math.Pi*base*t) * env * 0.34
		low := math.Sin(2*math.Pi*(base*0.5)*t) * decay(t, 0.85) * 0.2
		noise := bandNoise(t, 180, 1100, 0.45) * decay(t, 0.4) * 0.18
		return pan(main+low+noise, 0)
	}))
}

func render(seconds float64, fn func(t float64) stereoSample) []stereoSample {
	n := int(seconds * sampleRate)
	out := make([]stereoSample, 0, n)
	for i := 0; i < n; i++ {
		t := float64(i) / sampleRate
		s := fn(t)
		out = append(out, s)
	}
	return out
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

func clamp(v float64) float64 {
	if v > 1 {
		return 1
	}
	if v < -1 {
		return -1
	}
	return v
}

func decay(t, tau float64) float64 {
	if t < 0 {
		return 0
	}
	return math.Exp(-t / tau)
}

func attackDecay(t, attack, tau float64) float64 {
	if t < 0 {
		return 0
	}
	a := t / attack
	if a > 1 {
		a = 1
	}
	return a * math.Exp(-t/tau)
}

func bandNoise(t, lowHz, highHz, amp float64) float64 {
	n := math.Sin(2*math.Pi*lowHz*t+math.Sin(t*17.0)*0.7)*0.42 +
		math.Sin(2*math.Pi*((lowHz+highHz)/2)*t+math.Sin(t*11.0)*0.6)*0.33 +
		math.Sin(2*math.Pi*highHz*t+math.Sin(t*7.3)*0.4)*0.25
	n += (rand.Float64()*2 - 1) * 0.12
	return n * amp
}

func pan(v, p float64) stereoSample {
	if p > 1 {
		p = 1
	}
	if p < -1 {
		p = -1
	}
	l := v * (1 - p*0.35)
	r := v * (1 + p*0.35)
	return stereoSample{l: l, r: r}
}

func jitterPan(t float64) float64 {
	return math.Sin(t*19.0)*0.5 + math.Sin(t*7.0)*0.3
}
