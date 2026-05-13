package whisper

import (
	"os"
	"testing"
	"unsafe"

	"github.com/ardanlabs/bucky/pkg/audio"
)

// TestVadParamsSize verifies the Go mirror of whisper_vad_params has the
// expected size and that defaults round-trip cleanly through the FFI
// boundary.
func TestVadParamsSize(t *testing.T) {
	testSetup(t)

	if got := unsafe.Sizeof(VadParams{}); got != 24 {
		t.Errorf("unsafe.Sizeof(VadParams) = %d, want 24", got)
	}

	p := VadDefaultParams()

	// Whisper.cpp ships sane defaults: positive threshold, positive
	// min/max durations, non-negative pad/overlap.
	if p.Threshold <= 0 || p.Threshold > 1 {
		t.Errorf("Threshold = %v, want in (0,1]", p.Threshold)
	}
	if p.MinSpeechDurationMs <= 0 {
		t.Errorf("MinSpeechDurationMs = %d, want > 0", p.MinSpeechDurationMs)
	}
	if p.MinSilenceDurationMs <= 0 {
		t.Errorf("MinSilenceDurationMs = %d, want > 0", p.MinSilenceDurationMs)
	}
	if p.MaxSpeechDurationS <= 0 {
		t.Errorf("MaxSpeechDurationS = %v, want > 0", p.MaxSpeechDurationS)
	}
	if p.SpeechPadMs < 0 {
		t.Errorf("SpeechPadMs = %d, want >= 0", p.SpeechPadMs)
	}
	if p.SamplesOverlap < 0 {
		t.Errorf("SamplesOverlap = %v, want >= 0", p.SamplesOverlap)
	}
}

// TestVadContextParamsSize verifies the Go mirror of
// whisper_vad_context_params has the expected size and defaults.
func TestVadContextParamsSize(t *testing.T) {
	testSetup(t)

	if got := unsafe.Sizeof(VadContextParams{}); got != 12 {
		t.Errorf("unsafe.Sizeof(VadContextParams) = %d, want 12", got)
	}

	p := VadDefaultContextParams()
	if p.NThreads <= 0 {
		t.Errorf("NThreads = %d, want > 0", p.NThreads)
	}
	// Just ensure these fields can be read without crashing.
	_ = p.UseGPU
	_ = p.GPUDevice
}

// TestVadDetectsSpeechInJFK exercises the full VAD pipeline on the bundled
// JFK sample. Requires BUCKY_VAD_MODEL pointing at a ggml-silero-v5.1.2 (or
// similar) GGUF VAD model.
func TestVadDetectsSpeechInJFK(t *testing.T) {
	testSetup(t)

	vadModel := os.Getenv("BUCKY_VAD_MODEL")
	if vadModel == "" {
		t.Skip("BUCKY_VAD_MODEL not set; skipping VAD pipeline test")
	}
	if _, err := os.Stat(vadModel); err != nil {
		t.Skipf("VAD model %q not present: %v", vadModel, err)
	}

	audioPath := testAudioFileName(t)

	cparams := VadDefaultContextParams()
	vctx, err := VadInitFromFileWithParams(vadModel, cparams)
	if err != nil {
		t.Fatalf("VadInitFromFileWithParams: %v", err)
	}
	defer VadFree(vctx)

	f, err := os.Open(audioPath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer f.Close()
	samples, err := audio.Decode(f)
	if err != nil {
		t.Fatalf("audio.Decode: %v", err)
	}

	if !VadDetectSpeech(vctx, samples) {
		t.Fatal("VadDetectSpeech returned false")
	}
	if got := VadNProbs(vctx); got <= 0 {
		t.Fatalf("VadNProbs = %d, want > 0", got)
	}

	params := VadDefaultParams()
	segs, err := VadSegmentsFromProbs(vctx, params)
	if err != nil {
		t.Fatalf("VadSegmentsFromProbs: %v", err)
	}
	defer VadFreeSegments(segs)

	n := VadSegmentsNSegments(segs)
	if n <= 0 {
		t.Fatalf("VadSegmentsNSegments = %d, want > 0", n)
	}
	for i := int32(0); i < n; i++ {
		t0 := VadSegmentsGetSegmentT0(segs, i)
		t1 := VadSegmentsGetSegmentT1(segs, i)
		t.Logf("vad segment %d: [%.2fs -> %.2fs]", i, t0/100, t1/100)
		if t1 < t0 {
			t.Errorf("segment %d: t1 (%v) < t0 (%v)", i, t1, t0)
		}
	}
}
