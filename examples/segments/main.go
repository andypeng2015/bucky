// segments transcribes an audio clip and prints each segment with its
// start and end timestamps. Optionally prints per-token detail with -tokens.
//
// Usage:
//
//	BUCKY_LIB=./lib BUCKY_TEST_MODEL=$HOME/models/ggml-tiny.bin \
//	    go run ./examples/segments samples/jfk.wav
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ardanlabs/bucky/pkg/audio"
	"github.com/ardanlabs/bucky/pkg/whisper"
)

func main() {
	var (
		showTokens = flag.Bool("tokens", false, "print per-token id, text, and probability")
		threads    = flag.Int("threads", 4, "number of CPU threads")
	)
	flag.Parse()
	if flag.NArg() < 1 {
		log.Fatalf("usage: %s [flags] <audio-file>", os.Args[0])
	}
	audioPath := flag.Arg(0)

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
	wparams.NThreads = int32(*threads)
	wparams.PrintProgress = 0
	wparams.PrintRealtime = 0
	wparams.PrintTimestamps = 0
	if *showTokens {
		wparams.TokenTimestamps = 1
	}

	if err := whisper.Full(ctx, wparams, samples); err != nil {
		log.Fatalf("Full: %v", err)
	}

	for i := int32(0); i < whisper.FullNSegments(ctx); i++ {
		t0 := whisper.FullGetSegmentT0(ctx, i) * 10
		t1 := whisper.FullGetSegmentT1(ctx, i) * 10
		text := whisper.FullGetSegmentText(ctx, i)
		fmt.Printf("[%s -> %s] %s\n", formatMs(t0), formatMs(t1), text)

		if *showTokens {
			for j := int32(0); j < whisper.FullNTokens(ctx, i); j++ {
				td := whisper.FullGetTokenData(ctx, i, j)
				txt := whisper.FullGetTokenText(ctx, i, j)
				fmt.Printf("  token[%d] id=%d p=%.3f text=%q\n", j, td.ID, td.P, txt)
			}
		}
	}
}

func formatMs(ms int64) string {
	s := ms / 1000
	mm := s / 60
	ss := s % 60
	return fmt.Sprintf("%02d:%02d.%03d", mm, ss, ms%1000)
}
