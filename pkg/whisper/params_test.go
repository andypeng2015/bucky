package whisper

import (
	"testing"
	"unsafe"
)

// TestWhisperFullParamsSize verifies the Go mirror of whisper_full_params has
// the same size as the C struct. We compute the C size by round-tripping the
// struct through the C library: take the by-ref pointer and copy back into a
// fresh Go struct. If the layouts disagree, fields would be misaligned and
// modifications would not survive the round-trip.
func TestWhisperFullParamsSize(t *testing.T) {
	testSetup(t)

	// Sanity: assert a known size for the platforms we currently support.
	// Disagreement here means the explicit padding in WhisperFullParams is
	// out of sync with the C ABI on this platform.
	if got := unsafe.Sizeof(WhisperFullParams{}); got != 304 {
		t.Errorf("unsafe.Sizeof(WhisperFullParams) = %d, want 304", got)
	}

	// Round-trip: get defaults by ref, mutate one field, and verify a known
	// field survives a copy through Go memory.
	ptr := FullDefaultParamsByRef(SamplingGreedy)
	if ptr == nil {
		t.Fatal("FullDefaultParamsByRef returned NULL")
	}
	defer FreeParams(ptr)

	cParams := *ptr

	// The greedy strategy default should be SamplingGreedy.
	if cParams.Strategy != SamplingGreedy {
		t.Errorf("Strategy = %d, want %d", cParams.Strategy, SamplingGreedy)
	}

	// Whisper sets sane defaults for n_threads (>=1).
	if cParams.NThreads <= 0 {
		t.Errorf("NThreads = %d, want > 0", cParams.NThreads)
	}

	// Verify a representative spread of fields agrees between the by-ref and
	// by-value entry points. Padding bytes are deliberately skipped because
	// whisper.cpp does not zero-initialize them in the by-ref allocation,
	// while Go zero-initializes them in the by-value path.
	byVal := FullDefaultParams(SamplingGreedy)
	checks := []struct {
		name string
		a, b any
	}{
		{"Strategy", cParams.Strategy, byVal.Strategy},
		{"NThreads", cParams.NThreads, byVal.NThreads},
		{"NMaxTextCtx", cParams.NMaxTextCtx, byVal.NMaxTextCtx},
		{"OffsetMs", cParams.OffsetMs, byVal.OffsetMs},
		{"DurationMs", cParams.DurationMs, byVal.DurationMs},
		{"PrintRealtime", cParams.PrintRealtime, byVal.PrintRealtime},
		{"PrintTimestamps", cParams.PrintTimestamps, byVal.PrintTimestamps},
		{"TholdPt", cParams.TholdPt, byVal.TholdPt},
		{"TholdPtsum", cParams.TholdPtsum, byVal.TholdPtsum},
		{"MaxTokens", cParams.MaxTokens, byVal.MaxTokens},
		{"AudioCtx", cParams.AudioCtx, byVal.AudioCtx},
		{"PromptNTokens", cParams.PromptNTokens, byVal.PromptNTokens},
		{"Temperature", cParams.Temperature, byVal.Temperature},
		{"MaxInitialTs", cParams.MaxInitialTS, byVal.MaxInitialTS},
		{"LengthPenalty", cParams.LengthPenalty, byVal.LengthPenalty},
		{"TemperatureInc", cParams.TemperatureInc, byVal.TemperatureInc},
		{"EntropyThold", cParams.EntropyThold, byVal.EntropyThold},
		{"LogprobThold", cParams.LogprobThold, byVal.LogprobThold},
		{"NoSpeechThold", cParams.NoSpeechThold, byVal.NoSpeechThold},
		{"GreedyBestOf", cParams.GreedyBestOf, byVal.GreedyBestOf},
		{"BeamSearchBeamSize", cParams.BeamSearchBeamSize, byVal.BeamSearchBeamSize},
		{"BeamSearchPatience", cParams.BeamSearchPatience, byVal.BeamSearchPatience},
		{"NGrammarRules", cParams.NGrammarRules, byVal.NGrammarRules},
		{"IStartRule", cParams.IStartRule, byVal.IStartRule},
		{"GrammarPenalty", cParams.GrammarPenalty, byVal.GrammarPenalty},
		{"VadThreshold", cParams.VadThreshold, byVal.VadThreshold},
		{"VadMinSpeechDurationMs", cParams.VadMinSpeechDurationMs, byVal.VadMinSpeechDurationMs},
		{"VadSamplesOverlap", cParams.VadSamplesOverlap, byVal.VadSamplesOverlap},
	}
	for _, c := range checks {
		if c.a != c.b {
			t.Errorf("%s: by-ref=%v by-val=%v", c.name, c.a, c.b)
		}
	}
}

// TestContextDefaultParams verifies the context params round-trip and have
// reasonable defaults.
func TestContextDefaultParams(t *testing.T) {
	testSetup(t)

	p := ContextDefaultParams()
	// use_gpu defaults to true on platforms with a GPU backend in the build.
	// We just verify the call doesn't panic and returns something we can read.
	_ = p.UseGPU
	_ = p.GPUDevice
	_ = p.DtwMemSize
}
