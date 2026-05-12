// hello is the smallest possible bucky example: load a tiny whisper model,
// transcribe a 16 kHz mono 16-bit PCM WAV file, and print the resulting text.
//
// Usage:
//
//	BUCKY_LIB=./lib BUCKY_TEST_MODEL=$HOME/models/ggml-tiny.bin \
//	    go run ./examples/hello samples/jfk.wav
package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/ardanlabs/bucky/pkg/whisper"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("usage: %s <wav-file>", os.Args[0])
	}
	wavPath := os.Args[1]

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

	cparams := whisper.ContextDefaultParams()
	ctx, err := whisper.InitFromFileWithParams(modelPath, cparams)
	if err != nil {
		log.Fatalf("InitFromFileWithParams: %v", err)
	}
	defer whisper.Free(ctx)

	samples, err := loadWAV16kMono(wavPath)
	if err != nil {
		log.Fatalf("loadWAV16kMono: %v", err)
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

// loadWAV16kMono reads a 16 kHz mono 16-bit PCM WAV file into a float32 PCM
// slice in the [-1, 1] range. This is intentionally minimal; full-fledged
// audio decoding lives in pkg/audio (PR #3).
func loadWAV16kMono(path string) ([]float32, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var hdr struct {
		Riff      [4]byte
		ChunkSize uint32
		Wave      [4]byte
	}
	if err := binary.Read(f, binary.LittleEndian, &hdr); err != nil {
		return nil, err
	}
	if string(hdr.Riff[:]) != "RIFF" || string(hdr.Wave[:]) != "WAVE" {
		return nil, errors.New("not a RIFF/WAVE file")
	}

	var (
		fmtFound  bool
		dataFound bool
		channels  uint16
		sampleRt  uint32
		bitsPer   uint16
		samples   []float32
	)

	for !dataFound {
		var sub struct {
			Id   [4]byte
			Size uint32
		}
		if err := binary.Read(f, binary.LittleEndian, &sub); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		switch string(sub.Id[:]) {
		case "fmt ":
			var fmtChunk struct {
				AudioFormat   uint16
				NumChannels   uint16
				SampleRate    uint32
				ByteRate      uint32
				BlockAlign    uint16
				BitsPerSample uint16
			}
			if err := binary.Read(f, binary.LittleEndian, &fmtChunk); err != nil {
				return nil, err
			}
			if fmtChunk.AudioFormat != 1 {
				return nil, errors.New("only PCM WAV is supported")
			}
			channels = fmtChunk.NumChannels
			sampleRt = fmtChunk.SampleRate
			bitsPer = fmtChunk.BitsPerSample
			fmtFound = true
			if extra := int64(sub.Size) - 16; extra > 0 {
				if _, err := f.Seek(extra, io.SeekCurrent); err != nil {
					return nil, err
				}
			}
		case "data":
			if !fmtFound {
				return nil, errors.New("data chunk before fmt chunk")
			}
			if channels != 1 || sampleRt != whisper.SampleRate || bitsPer != 16 {
				return nil, fmt.Errorf("expected 16 kHz mono 16-bit PCM WAV (got %d Hz, %d ch, %d bit)", sampleRt, channels, bitsPer)
			}
			n := int(sub.Size) / 2
			raw := make([]int16, n)
			if err := binary.Read(f, binary.LittleEndian, raw); err != nil {
				return nil, err
			}
			samples = make([]float32, n)
			for i, v := range raw {
				samples[i] = float32(v) / 32768.0
			}
			dataFound = true
		default:
			if _, err := f.Seek(int64(sub.Size), io.SeekCurrent); err != nil {
				return nil, err
			}
		}
	}

	if !dataFound {
		return nil, errors.New("no data chunk in WAV")
	}
	return samples, nil
}
