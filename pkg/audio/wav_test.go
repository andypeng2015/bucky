package audio

import (
	"bytes"
	"encoding/binary"
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
