// words transcribes an audio clip with token-level timestamps enabled and
// prints "[t0 -> t1] word" for every non-special token. This is the basic
// (whisper.cpp built-in) word-timing path.
//
// Usage:
//
//	BUCKY_LIB=./lib BUCKY_TEST_MODEL=$HOME/models/ggml-tiny.bin \
//	    go run ./examples/words samples/jfk.wav
//
// There is also an experimental DTW (Dynamic Time Warping) path in
// whisper.cpp that produces tighter word-level alignments. To use it, set
// these BEFORE calling InitFromFileWithParams (DTW heads are baked into the
// context at init time, not toggled per Full call):
//
//	cparams := whisper.ContextDefaultParams()
//	cparams.DtwTokenTimestamps = 1
//	cparams.DtwAheadsPreset = whisper.AHeadsBaseEN // matches the loaded model
//	cparams.DtwMemSize = 128 * 1024 * 1024         // 128 MiB scratch
//	ctx, _ := whisper.InitFromFileWithParams(modelPath, cparams)
//
// The TokenData.TDtw field is then populated alongside the regular T0/T1.
// Defaulting to plain TokenTimestamps here because DTW requires choosing the
// right alignment-heads preset per model and adds memory pressure.
package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ardanlabs/bucky/pkg/audio"
	"github.com/ardanlabs/bucky/pkg/whisper"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("usage: %s <audio-file>", os.Args[0])
	}
	audioPath := os.Args[1]

	libPath := os.Getenv("BUCKY_LIB")
	if libPath == "" {
		log.Fatal("BUCKY_LIB must point to the directory containing libwhisper")
	}
	modelPath := os.Getenv("BUCKY_TEST_MODEL")
	if modelPath == "" {
		log.Fatal("BUCKY_TEST_MODEL must point to a GGML whisper model")
	}

	if err := whisper.Load(libPath); err != nil {
		log.Fatalf("whisper.Load: %v", err)
	}

	cparams := whisper.ContextDefaultParams()
	ctx, err := whisper.InitFromFileWithParams(modelPath, cparams)
	if err != nil {
		log.Fatalf("InitFromFileWithParams: %v", err)
	}
	defer whisper.Free(ctx)

	f, err := os.Open(audioPath)
	if err != nil {
		log.Fatalf("open %s: %v", audioPath, err)
	}
	defer f.Close()
	samples, err := audio.Decode(f)
	if err != nil {
		log.Fatalf("audio.Decode: %v", err)
	}

	wparams := whisper.FullDefaultParams(whisper.SamplingGreedy)
	wparams.PrintProgress = 0
	wparams.PrintRealtime = 0
	wparams.PrintTimestamps = 0
	wparams.TokenTimestamps = 1 // enable per-token T0/T1 in TokenData

	if err := whisper.Full(ctx, wparams, samples); err != nil {
		log.Fatalf("Full: %v", err)
	}

	for i := int32(0); i < whisper.FullNSegments(ctx); i++ {
		for j := int32(0); j < whisper.FullNTokens(ctx, i); j++ {
			td := whisper.FullGetTokenData(ctx, i, j)
			text := whisper.FullGetTokenText(ctx, i, j)

			// Skip whisper's special control tokens (e.g. <|0.00|>,
			// <|notimestamps|>, <|en|>, <|transcribe|>). They are
			// tagged with the token text starting with "[" or "<|".
			trimmed := strings.TrimSpace(text)
			if trimmed == "" || strings.HasPrefix(trimmed, "[_") || strings.HasPrefix(trimmed, "<|") {
				continue
			}

			t0ms := td.T0 * 10
			t1ms := td.T1 * 10
			fmt.Printf("[%s -> %s] %s\n", formatMs(t0ms), formatMs(t1ms), text)
		}
	}
}

func formatMs(ms int64) string {
	if ms < 0 {
		ms = 0
	}
	s := ms / 1000
	mm := s / 60
	ss := s % 60
	return fmt.Sprintf("%02d:%02d.%03d", mm, ss, ms%1000)
}
