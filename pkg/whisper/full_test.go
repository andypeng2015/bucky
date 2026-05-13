package whisper

import (
	"os"
	"strings"
	"testing"

	"github.com/ardanlabs/bucky/pkg/audio"
)

func TestFullTranscribe(t *testing.T) {
	testSetup(t)
	modelPath := testModelFileName(t)
	audioPath := testAudioFileName(t)

	cparams := ContextDefaultParams()
	ctx, err := InitFromFileWithParams(modelPath, cparams)
	if err != nil {
		t.Fatalf("InitFromFileWithParams: %v", err)
	}
	defer Free(ctx)

	f, err := os.Open(audioPath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer f.Close()
	samples, err := audio.Decode(f)
	if err != nil {
		t.Fatalf("audio.Decode: %v", err)
	}
	if len(samples) == 0 {
		t.Fatal("audio.Decode returned no samples")
	}

	wparams := FullDefaultParams(SamplingGreedy)
	wparams.PrintProgress = 0
	wparams.PrintRealtime = 0
	wparams.PrintTimestamps = 0
	wparams.NoTimestamps = 1
	wparams.SingleSegment = 1

	if err := Full(ctx, wparams, samples); err != nil {
		t.Fatalf("Full: %v", err)
	}

	n := FullNSegments(ctx)
	if n <= 0 {
		t.Fatalf("FullNSegments = %d, want > 0", n)
	}

	var sb strings.Builder
	for i := int32(0); i < n; i++ {
		sb.WriteString(FullGetSegmentText(ctx, i))
	}
	got := strings.ToLower(sb.String())
	t.Logf("transcribed: %q", got)

	// Loose substring check: jfk.wav should mention "ask" somewhere.
	if !strings.Contains(got, "ask") {
		t.Errorf("transcription %q does not contain expected substring", got)
	}
}
