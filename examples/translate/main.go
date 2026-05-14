// translate transcribes an audio clip and translates the result to English.
//
// Usage:
//
//	BUCKY_LIB=./lib BUCKY_TEST_MODEL=$HOME/models/ggml-tiny.bin \
//	    go run ./examples/translate \
//			-lang es \
//			samples/spanish.mp3
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ardanlabs/bucky/pkg/audio"
	"github.com/ardanlabs/bucky/pkg/whisper"
)

func main() {
	var (
		lang    = flag.String("lang", "", "source language code (e.g. \"es\"); empty = auto-detect")
		threads = flag.Int("threads", 4, "number of CPU threads")
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
		log.Fatal("BUCKY_TEST_MODEL must point to a GGML whisper model (multilingual; not -en)")
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

	if !whisper.IsMultilingual(ctx) {
		log.Fatal("loaded model is not multilingual; use a non-`.en` whisper model for translation")
	}

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
	wparams.Translate = 1
	wparams.PrintProgress = 0
	wparams.PrintRealtime = 0
	wparams.PrintTimestamps = 0
	wparams.NoTimestamps = 1

	var refs whisper.StringRefs
	if err := refs.SetLanguage(&wparams, *lang); err != nil {
		log.Fatalf("SetLanguage: %v", err)
	}
	defer refs.KeepAlive()

	if err := whisper.Full(ctx, wparams, samples); err != nil {
		log.Fatalf("Full: %v", err)
	}

	var sb strings.Builder
	for i := int32(0); i < whisper.FullNSegments(ctx); i++ {
		sb.WriteString(whisper.FullGetSegmentText(ctx, i))
	}
	fmt.Println(strings.TrimSpace(sb.String()))
}
