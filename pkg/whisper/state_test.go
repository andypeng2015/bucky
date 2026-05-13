package whisper

import (
	"os"
	"strings"
	"testing"

	"github.com/ardanlabs/bucky/pkg/audio"
)

// TestFullWithState runs the same audio through both Full (default state) and
// FullWithState (explicit state) and verifies they return the same segment
// count and text. This exercises the whisper_state lifecycle and the
// _from_state accessors.
func TestFullWithState(t *testing.T) {
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

	mkParams := func() WhisperFullParams {
		p := FullDefaultParams(SamplingGreedy)
		p.PrintProgress = 0
		p.PrintRealtime = 0
		p.PrintTimestamps = 0
		p.NoTimestamps = 1
		p.SingleSegment = 1
		return p
	}

	// Run #1: default state via Full.
	if err := Full(ctx, mkParams(), samples); err != nil {
		t.Fatalf("Full: %v", err)
	}
	wantN := FullNSegments(ctx)
	if wantN <= 0 {
		t.Fatalf("FullNSegments = %d, want > 0", wantN)
	}
	var wantSb strings.Builder
	for i := int32(0); i < wantN; i++ {
		wantSb.WriteString(FullGetSegmentText(ctx, i))
	}
	wantText := wantSb.String()

	// Run #2: explicit state via FullWithState.
	state, err := InitState(ctx)
	if err != nil {
		t.Fatalf("InitState: %v", err)
	}
	defer FreeState(state)

	if err := FullWithState(ctx, state, mkParams(), samples); err != nil {
		t.Fatalf("FullWithState: %v", err)
	}
	gotN := FullNSegmentsFromState(state)
	if gotN != wantN {
		t.Errorf("FullNSegmentsFromState = %d, want %d (matches default-state)", gotN, wantN)
	}
	var gotSb strings.Builder
	for i := int32(0); i < gotN; i++ {
		gotSb.WriteString(FullGetSegmentTextFromState(state, i))
	}
	gotText := gotSb.String()
	if gotText != wantText {
		t.Errorf("FullWithState text mismatch:\n got=%q\nwant=%q", gotText, wantText)
	}

	// Exercise the remaining _from_state accessors at least once.
	if gotN > 0 {
		_ = FullGetSegmentT0FromState(state, 0)
		_ = FullGetSegmentT1FromState(state, 0)
		_ = FullGetSegmentNoSpeechProbFromState(state, 0)
		_ = FullGetSegmentSpeakerTurnNextFromState(state, 0)
		if nTokens := FullNTokensFromState(state, 0); nTokens > 0 {
			_ = FullGetTokenTextFromState(ctx, state, 0, 0)
			_ = FullGetTokenIDFromState(state, 0, 0)
			_ = FullGetTokenDataFromState(state, 0, 0)
			_ = FullGetTokenPFromState(state, 0, 0)
		}
	}
	_ = FullLangIDFromState(state)
}
