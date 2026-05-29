package audio

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math"
	"os"
	"testing"
)

// makeWAV synthesizes a minimal RIFF/WAVE file in memory with the given
// 16-bit PCM samples. Used to exercise the decoder without external fixtures.
func makeWAV(t *testing.T, sampleRate, channels int, samples []int16) []byte {
	t.Helper()
	var buf bytes.Buffer
	dataLen := uint32(len(samples) * 2)
	bytesPerSample := uint16(2)
	blockAlign := uint16(channels) * bytesPerSample
	byteRate := uint32(sampleRate) * uint32(blockAlign)

	buf.WriteString("RIFF")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(36+dataLen))
	buf.WriteString("WAVE")
	buf.WriteString("fmt ")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(16))
	_ = binary.Write(&buf, binary.LittleEndian, uint16(1)) // PCM
	_ = binary.Write(&buf, binary.LittleEndian, uint16(channels))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(sampleRate))
	_ = binary.Write(&buf, binary.LittleEndian, byteRate)
	_ = binary.Write(&buf, binary.LittleEndian, blockAlign)
	_ = binary.Write(&buf, binary.LittleEndian, uint16(16))
	buf.WriteString("data")
	_ = binary.Write(&buf, binary.LittleEndian, dataLen)
	for _, s := range samples {
		_ = binary.Write(&buf, binary.LittleEndian, s)
	}
	return buf.Bytes()
}

func TestDecodeWAVMono16(t *testing.T) {
	in := []int16{0, 16384, -16384, 32767, -32768}
	wav := makeWAV(t, 16000, 1, in)

	got, sr, ch, err := DecodeWAV(bytes.NewReader(wav))
	if err != nil {
		t.Fatalf("DecodeWAV: %v", err)
	}
	if sr != 16000 || ch != 1 {
		t.Fatalf("sampleRate=%d channels=%d, want 16000/1", sr, ch)
	}
	want := []float32{0, 16384.0 / 32768.0, -16384.0 / 32768.0, 32767.0 / 32768.0, -1.0}
	if len(got) != len(want) {
		t.Fatalf("len=%d, want %d", len(got), len(want))
	}
	for i := range want {
		if math.Abs(float64(got[i]-want[i])) > 1e-6 {
			t.Errorf("got[%d]=%f want %f", i, got[i], want[i])
		}
	}
}

func TestDecodeWAVStereo16Resample(t *testing.T) {
	// 4 frames @ 32000 Hz stereo. After downmix + resample to 16 kHz that's
	// 2 mono samples.
	in := []int16{1000, -1000, 2000, -2000, 3000, -3000, 4000, -4000}
	wav := makeWAV(t, 32000, 2, in)

	got, err := Decode(bytes.NewReader(wav))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len=%d, want 2", len(got))
	}
}

func TestDecodeRoundTripJFK(t *testing.T) {
	// The bundled JFK sample is 16 kHz mono 16-bit PCM.
	path := "../../samples/jfk.wav"
	data, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("samples/jfk.wav not present: %v", err)
	}
	got, err := Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("Decode returned no samples")
	}
	// jfk.wav is 11 seconds at 16 kHz → 176k samples.
	if len(got) < 100000 || len(got) > 200000 {
		t.Errorf("samples = %d, expected 100k–200k", len(got))
	}
}

// TestDecodeMP3Spanish exercises the MP3 decoder against the bundled
// 44.1 kHz stereo MP3 sample. Decode runs the full pipeline (sniff →
// DecodeMP3 → downmix → resample to 16 kHz mono), so this test
// transitively covers the MP3-detection branch in DecodeRawInto and the
// looksLikeMP3 helper.
func TestDecodeMP3Spanish(t *testing.T) {
	path := "../../samples/spanish.mp3"
	data, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("samples/spanish.mp3 not present: %v", err)
	}

	got, err := Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("Decode returned no samples")
	}

	// Loose sanity-check on the audio range: any non-trivial speech sample
	// should have at least some pixels above 1% of full scale. An MP3
	// decode that returned all zeros (decoder silently dropped frames)
	// would slip past a "len(got) > N" check otherwise.
	var maxAbs float32
	for _, v := range got {
		if v < 0 {
			v = -v
		}
		if v > maxAbs {
			maxAbs = v
		}
	}
	if maxAbs < 0.01 {
		t.Errorf("max|sample| = %v, expected speech-level audio above 0.01", maxAbs)
	}
}

// TestDecodeRawMP3 verifies the raw (pre-downmix, pre-resample) MP3 path
// reports a stereo stream at the source sample rate. go-mp3 always emits
// 16-bit little-endian stereo at the source rate; if that contract ever
// changes we want the Decode pipeline's downmix/resample stages to be
// audited.
func TestDecodeRawMP3(t *testing.T) {
	path := "../../samples/spanish.mp3"
	data, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("samples/spanish.mp3 not present: %v", err)
	}

	samples, sampleRate, channels, err := DecodeRaw(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("DecodeRaw: %v", err)
	}
	if channels != 2 {
		t.Errorf("channels = %d, want 2 (go-mp3 always emits stereo)", channels)
	}
	if sampleRate != 44100 {
		t.Errorf("sampleRate = %d, want 44100 (matches spanish.mp3 source)", sampleRate)
	}
	if len(samples) == 0 {
		t.Fatal("DecodeRaw returned no samples")
	}
}

// TestDecodeFLAC exercises the FLAC decoder against the bundled jfk.flac
// fixture (16 kHz mono, transcoded from samples/jfk.wav). Decode's sniff
// branch routes to DecodeFLAC, so this test exercises both the format
// detection and the FLAC decoding path.
func TestDecodeFLAC(t *testing.T) {
	path := "../../samples/jfk.flac"
	data, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("samples/jfk.flac not present: %v", err)
	}

	got, err := Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if len(got) < 100000 || len(got) > 200000 {
		t.Errorf("samples = %d, expected 100k–200k (~11s of 16 kHz audio)", len(got))
	}
}

// TestDecodeRawFLAC verifies the raw FLAC path reports the on-disk sample
// rate and channel count (not the post-downmix/resample values), and that
// samples are normalized to [-1, 1].
func TestDecodeRawFLAC(t *testing.T) {
	path := "../../samples/jfk.flac"
	data, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("samples/jfk.flac not present: %v", err)
	}

	samples, sampleRate, channels, err := DecodeRaw(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("DecodeRaw: %v", err)
	}
	if sampleRate != 16000 {
		t.Errorf("sampleRate = %d, want 16000 (jfk.flac is 16 kHz mono)", sampleRate)
	}
	if channels != 1 {
		t.Errorf("channels = %d, want 1", channels)
	}
	for i, v := range samples {
		if v < -1 || v > 1 {
			t.Fatalf("samples[%d] = %v, want value in [-1, 1]", i, v)
		}
	}
}

// TestDecodeUnsupportedFormat verifies the format-sniffing dispatch
// surfaces a useful, well-formed error for unrecognized inputs.
func TestDecodeUnsupportedFormat(t *testing.T) {
	_, err := Decode(bytes.NewReader([]byte("not an audio file at all")))
	if err == nil {
		t.Fatal("Decode: expected error for unknown magic, got nil")
	}
	if !errors.Is(err, ErrUnsupportedFormat) {
		t.Errorf("expected errors.Is(ErrUnsupportedFormat), got %v", err)
	}
}
