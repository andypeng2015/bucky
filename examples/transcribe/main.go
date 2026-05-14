// transcribe is a fuller bucky example: it accepts CLI flags to set the
// language, initial prompt, temperature, and beam size when transcribing an
// audio file with whisper.
//
// Usage:
//
//	BUCKY_LIB=./lib BUCKY_TEST_MODEL=$HOME/models/ggml-tiny.bin \
//	    go run ./examples/transcribe \
//	        -lang es \
//	        -prompt "Woman Talking" \
//	        samples/spanish.mp3
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
		lang        = flag.String("lang", "", "language code (e.g. \"en\"); empty = auto-detect")
		prompt      = flag.String("prompt", "", "initial prompt text to bias the decoder")
		temperature = flag.Float64("temperature", 0.0, "decoding temperature (0 = greedy)")
		beamSize    = flag.Int("beam", 0, "beam size (>0 enables beam-search sampling)")
		threads     = flag.Int("threads", 4, "number of CPU threads")
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

	strategy := whisper.SamplingGreedy
	if *beamSize > 0 {
		strategy = whisper.SamplingBeamSearch
	}
	wparams := whisper.FullDefaultParams(strategy)
	wparams.NThreads = int32(*threads)
	wparams.Temperature = float32(*temperature)
	if *beamSize > 0 {
		wparams.BeamSearchBeamSize = int32(*beamSize)
	}
	wparams.PrintProgress = 0
	wparams.PrintRealtime = 0
	wparams.PrintTimestamps = 0
	wparams.NoTimestamps = 1

	var refs whisper.StringRefs
	if err := refs.SetLanguage(&wparams, *lang); err != nil {
		log.Fatalf("SetLanguage: %v", err)
	}
	if err := refs.SetInitialPrompt(&wparams, *prompt); err != nil {
		log.Fatalf("SetInitialPrompt: %v", err)
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

	if id := whisper.FullLangID(ctx); id >= 0 {
		fmt.Fprintf(os.Stderr, "(language: %s)\n", whisper.LangStr(id))
	}
}
