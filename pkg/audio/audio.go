// Package audio decodes common audio formats into the 16 kHz mono float32
// PCM that whisper.cpp expects. Decoders are pure Go (no CGo) and cover
// WAV (16/24/32-bit PCM and 32-bit float), MP3 (via hajimehoshi/go-mp3) and
// FLAC (via mewkiz/flac). m4a/ogg/webm are intentionally unsupported in v1.
package audio

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

// TargetSampleRate is the sample rate whisper.cpp expects (16 kHz).
const TargetSampleRate = 16000

// ErrUnsupportedFormat is returned when Decode cannot identify the input.
var ErrUnsupportedFormat = errors.New("audio: unsupported format")

// Decode sniffs the magic bytes of the input and decodes it into 16 kHz mono
// float32 PCM in the range [-1.0, 1.0]. The returned slice is ready to pass
// to whisper.Full.
func Decode(r io.Reader) ([]float32, error) {
	samples, sampleRate, channels, err := DecodeRaw(r)
	if err != nil {
		return nil, err
	}
	mono := DownmixToMono(samples, channels)
	if sampleRate != TargetSampleRate {
		mono = ResampleLinear(mono, sampleRate, TargetSampleRate)
	}
	return mono, nil
}

// DecodeRaw sniffs the magic bytes and dispatches to the appropriate
// format-specific decoder. It returns the samples in their native sample
// rate and channel layout (interleaved when channels > 1), as float32 in
// the range [-1.0, 1.0].
func DecodeRaw(r io.Reader) (samples []float32, sampleRate int, channels int, err error) {
	const sniffN = 12
	head := make([]byte, sniffN)
	n, rerr := io.ReadFull(r, head)
	if rerr != nil && !errors.Is(rerr, io.ErrUnexpectedEOF) && !errors.Is(rerr, io.EOF) {
		return nil, 0, 0, rerr
	}
	head = head[:n]
	combined := io.MultiReader(bytes.NewReader(head), r)

	switch {
	case n >= 12 && string(head[0:4]) == "RIFF" && string(head[8:12]) == "WAVE":
		return DecodeWAV(combined)
	case n >= 4 && string(head[0:4]) == "fLaC":
		return DecodeFLAC(combined)
	case looksLikeMP3(head):
		return DecodeMP3(combined)
	default:
		return nil, 0, 0, fmt.Errorf("%w: unknown magic %x", ErrUnsupportedFormat, head)
	}
}

// looksLikeMP3 returns true for streams that begin with an ID3v2 tag or an
// MPEG audio frame sync (11 set bits).
func looksLikeMP3(head []byte) bool {
	if len(head) >= 3 && string(head[0:3]) == "ID3" {
		return true
	}
	if len(head) >= 2 && head[0] == 0xFF && (head[1]&0xE0) == 0xE0 {
		return true
	}
	return false
}
