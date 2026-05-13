package whisper

import (
	"os"
	"testing"

	"github.com/ardanlabs/bucky/pkg/audio"
)

// BenchmarkFullJFK measures end-to-end transcription throughput on the JFK
// sample with greedy sampling.
//
// Requires: BUCKY_LIB and one of (BUCKY_BENCH_MODEL or BUCKY_TEST_MODEL).
// BUCKY_BENCH_MODEL takes precedence so callers can run the same benchmark
// across tiny/base/small without disturbing the regular test model.
//
// Reports the canonical bench timer (ns/op) and a custom "rtf" metric:
// real-time factor = wall_time / audio_duration. RTF < 1 is faster than real
// time.
func BenchmarkFullJFK(b *testing.B) {
	libPath := os.Getenv("BUCKY_LIB")
	if libPath == "" {
		b.Skip("BUCKY_LIB not set; skipping benchmark")
	}

	loadOnce.Do(func() { loadErr = Load(libPath) })
	if loadErr != nil {
		b.Fatalf("Load: %v", loadErr)
	}

	modelPath := os.Getenv("BUCKY_BENCH_MODEL")
	if modelPath == "" {
		modelPath = os.Getenv("BUCKY_TEST_MODEL")
	}
	if modelPath == "" {
		b.Skip("BUCKY_BENCH_MODEL / BUCKY_TEST_MODEL not set; skipping benchmark")
	}

	audioPath := os.Getenv("BUCKY_TEST_AUDIO")
	if audioPath == "" {
		b.Skip("BUCKY_TEST_AUDIO not set; skipping benchmark")
	}

	cparams := ContextDefaultParams()
	if os.Getenv("BUCKY_USE_GPU") == "0" {
		// CPU-only Linux artifacts assert when use_gpu=1; see
		// helpers_test.go testContextDefaultParams for the why.
		cparams.UseGPU = 0
	}
	ctx, err := InitFromFileWithParams(modelPath, cparams)
	if err != nil {
		b.Fatalf("InitFromFileWithParams: %v", err)
	}
	defer Free(ctx)

	f, err := os.Open(audioPath)
	if err != nil {
		b.Fatalf("open: %v", err)
	}
	defer f.Close()
	samples, err := audio.Decode(f)
	if err != nil {
		b.Fatalf("audio.Decode: %v", err)
	}
	if len(samples) == 0 {
		b.Fatal("no samples")
	}
	audioSeconds := float64(len(samples)) / float64(SampleRate)

	mkParams := func() WhisperFullParams {
		p := FullDefaultParams(SamplingGreedy)
		p.PrintProgress = 0
		p.PrintRealtime = 0
		p.PrintTimestamps = 0
		p.NoTimestamps = 1
		p.SingleSegment = 1
		return p
	}

	// Warm up: first run on Metal includes JIT/library setup we don't want
	// folded into the timed loop.
	if err := Full(ctx, mkParams(), samples); err != nil {
		b.Fatalf("Full (warmup): %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := Full(ctx, mkParams(), samples); err != nil {
			b.Fatalf("Full: %v", err)
		}
	}
	b.StopTimer()

	// Report real-time factor: wall seconds per second of audio.
	wallSeconds := b.Elapsed().Seconds() / float64(b.N)
	b.ReportMetric(wallSeconds/audioSeconds, "rtf")
	b.ReportMetric(audioSeconds, "audio_s")
}
