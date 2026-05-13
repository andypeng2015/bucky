package audio

// DownmixToMono averages all channels into a single mono channel. The input
// is assumed to be interleaved (L0,R0,L1,R1,...). When channels == 1 the
// input is returned as-is.
func DownmixToMono(samples []float32, channels int) []float32 {
	if channels <= 1 {
		return samples
	}
	frames := len(samples) / channels
	out := make([]float32, frames)
	inv := 1.0 / float32(channels)
	for i := range frames {
		var sum float32
		for c := range channels {
			sum += samples[i*channels+c]
		}
		out[i] = sum * inv
	}
	return out
}

// ResampleLinear converts samples from inRate to outRate using linear
// interpolation. The input is mono. Linear interpolation is fast and
// adequate for whisper input; it does no anti-alias filtering, so very
// high frequencies will alias when downsampling. Whisper itself is
// trained on 16 kHz so this is rarely audible in practice.
func ResampleLinear(samples []float32, inRate, outRate int) []float32 {
	if inRate == outRate || len(samples) == 0 {
		return samples
	}
	ratio := float64(inRate) / float64(outRate)
	outN := int(float64(len(samples)) / ratio)
	out := make([]float32, outN)
	for i := range outN {
		pos := float64(i) * ratio
		idx := int(pos)
		frac := float32(pos - float64(idx))
		if idx+1 >= len(samples) {
			out[i] = samples[len(samples)-1]
			continue
		}
		out[i] = samples[idx]*(1-frac) + samples[idx+1]*frac
	}
	return out
}
