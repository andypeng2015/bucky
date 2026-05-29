package cmd

import (
	"regexp"
	"strings"
	"testing"
)

// TestWhisperCatalogValid walks every catalog entry and verifies the
// shape downstream code relies on: well-formed URL on the canonical
// HuggingFace mirror, ending with .bin, with a human-readable size. A
// typo on a catalog edit otherwise lands in production unnoticed —
// `bucky model get <name>` would just return an HTTP 404.
func TestWhisperCatalogValid(t *testing.T) {
	if len(whisperCatalog) == 0 {
		t.Fatal("whisperCatalog is empty")
	}

	sizeRe := regexp.MustCompile(`^\d+(\.\d+)?\s+(KB|MB|GB)$`)

	for name, entry := range whisperCatalog {
		t.Run(name, func(t *testing.T) {
			if !strings.HasPrefix(entry.URL, "https://huggingface.co/") {
				t.Errorf("URL = %q, want https://huggingface.co/... prefix", entry.URL)
			}
			if !strings.HasSuffix(entry.URL, ".bin") {
				t.Errorf("URL = %q, want .bin suffix", entry.URL)
			}
			if !sizeRe.MatchString(entry.Size) {
				t.Errorf("Size = %q, want \"<number> (KB|MB|GB)\"", entry.Size)
			}
		})
	}
}

// TestWhisperModelByName verifies the lookup helper handles both the
// short name and the ggml-<name>.bin convenience form, and returns
// false on unknown names.
func TestWhisperModelByName(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"tiny", true},
		{"tiny.en", true},
		{"large-v3", true},
		{"silero-vad", true},

		// Convenience form: full ggml-*.bin filename round-trips back
		// to the short name in the catalog map.
		{"ggml-tiny.bin", true},
		{"ggml-large-v3-turbo.bin", true},

		{"", false},
		{"does-not-exist", false},
		{"ggml-does-not-exist.bin", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, ok := whisperModelByName(tt.input)
			if ok != tt.want {
				t.Errorf("whisperModelByName(%q) = %v, want %v", tt.input, ok, tt.want)
			}
		})
	}
}
