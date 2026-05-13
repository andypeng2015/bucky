package audio

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// BenchmarkDecodeWAV measures the WAV → []float32 path on the bundled JFK
// sample. The file is read into memory once so the benchmark times decoding,
// not disk I/O.
//
// Honors BUCKY_TEST_AUDIO when set; otherwise falls back to the vendored
// samples/jfk.wav at the repo root.
func BenchmarkDecodeWAV(b *testing.B) {
	data := loadAudioFixture(b)
	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, _, _, err := DecodeWAV(bytes.NewReader(data)); err != nil {
			b.Fatalf("DecodeWAV: %v", err)
		}
	}
}

// BenchmarkDecode measures the full sniff-then-decode-then-resample-to-mono
// path. With a 16 kHz mono input this collapses to DecodeWAV plus the
// downmix/resample no-ops, which is the realistic call shape used by the
// CLI and examples.
func BenchmarkDecode(b *testing.B) {
	data := loadAudioFixture(b)
	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := Decode(bytes.NewReader(data)); err != nil {
			b.Fatalf("Decode: %v", err)
		}
	}
}

// BenchmarkDecodeWAVInto measures the buffer-reuse fast path. The dst slice
// is grown once during the warm-up call and reused on every timed iteration,
// so the per-call []float32 allocation drops to zero.
func BenchmarkDecodeWAVInto(b *testing.B) {
	data := loadAudioFixture(b)
	b.SetBytes(int64(len(data)))

	// Warm up: grow dst to its full size so the timed loop exercises the
	// steady-state, capacity-already-sufficient path.
	dst, _, _, err := DecodeWAVInto(nil, bytes.NewReader(data))
	if err != nil {
		b.Fatalf("DecodeWAVInto warmup: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dst, _, _, err = DecodeWAVInto(dst[:0], bytes.NewReader(data))
		if err != nil {
			b.Fatalf("DecodeWAVInto: %v", err)
		}
	}
}

// BenchmarkDecodeInto measures the same buffer-reuse fast path through the
// sniff/dispatch wrapper used by the CLI and examples.
func BenchmarkDecodeInto(b *testing.B) {
	data := loadAudioFixture(b)
	b.SetBytes(int64(len(data)))

	dst, err := DecodeInto(nil, bytes.NewReader(data))
	if err != nil {
		b.Fatalf("DecodeInto warmup: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dst, err = DecodeInto(dst[:0], bytes.NewReader(data))
		if err != nil {
			b.Fatalf("DecodeInto: %v", err)
		}
	}
}

// loadAudioFixture returns the audio bytes for benchmarks. Callers don't
// need the disk read to be in the timed loop, so this is invoked before
// b.ResetTimer.
func loadAudioFixture(b *testing.B) []byte {
	b.Helper()

	if path := os.Getenv("BUCKY_TEST_AUDIO"); path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			b.Fatalf("read BUCKY_TEST_AUDIO=%s: %v", path, err)
		}
		return data
	}

	// Fall back to the repo-vendored sample. Test cwd is the package
	// directory (pkg/audio), so the JFK sample lives two levels up.
	fallback := filepath.Join("..", "..", "samples", "jfk.wav")
	data, err := os.ReadFile(fallback)
	if err != nil {
		b.Skipf("BUCKY_TEST_AUDIO not set and fallback %s missing: %v", fallback, err)
	}
	return data
}
