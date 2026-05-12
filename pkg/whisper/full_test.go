package whisper

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
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

	samples, err := loadWAV16kMono(audioPath)
	if err != nil {
		t.Fatalf("loadWAV16kMono: %v", err)
	}
	if len(samples) == 0 {
		t.Fatal("loadWAV16kMono returned no samples")
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

// loadWAV16kMono reads a minimal 16-bit PCM WAV file (16 kHz mono) and
// returns the samples as []float32 in the [-1.0, 1.0] range. Used only by
// tests; pkg/audio is PR #3.
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
			// Skip any extra fmt-chunk bytes.
			if extra := int64(sub.Size) - 16; extra > 0 {
				if _, err := f.Seek(extra, io.SeekCurrent); err != nil {
					return nil, err
				}
			}
		case "data":
			if !fmtFound {
				return nil, errors.New("data chunk before fmt chunk")
			}
			if channels != 1 || sampleRt != SampleRate || bitsPer != 16 {
				return nil, errors.New("expected 16 kHz mono 16-bit PCM WAV")
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
