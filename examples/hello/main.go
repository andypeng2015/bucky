// hello is the smallest possible bucky example: load a tiny whisper model,
// decode an audio file (WAV / MP3 / FLAC), and print the resulting text.
//
// Usage:
//
//	BUCKY_LIB=./lib BUCKY_TEST_MODEL=$HOME/models/ggml-tiny.bin \
//	    go run ./examples/hello samples/jfk.wav
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
		log.Fatal("BUCKY_TEST_MODEL must point to a GGML whisper model (e.g. ggml-tiny.bin)")
	}

	if err := whisper.Load(libPath); err != nil {
		log.Fatalf("whisper.Load: %v", err)
	}

	if err := whisper.Init(libPath); err != nil {
		log.Fatalf("whisper.Init: %v", err)
	}

	cparams := whisper.ContextDefaultParams()
	// Honor BUCKY_USE_GPU=0 so this example works against the CPU-only
	// Linux artifacts (which assert when use_gpu=1 with no GPU backend).
	if os.Getenv("BUCKY_USE_GPU") == "0" {
		cparams.UseGPU = 0
	}
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
	wparams.NoTimestamps = 1

	if err := whisper.Full(ctx, wparams, samples); err != nil {
		log.Fatalf("Full: %v", err)
	}

	var sb strings.Builder
	for i := int32(0); i < whisper.FullNSegments(ctx); i++ {
		sb.WriteString(whisper.FullGetSegmentText(ctx, i))
	}
	fmt.Println(strings.TrimSpace(sb.String()))
}
