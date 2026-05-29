package loader

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGetLibraryFilename(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		lib      string
		expected map[string]string // OS -> expected result
	}{
		{
			name: "whisper library",
			path: "/usr/local/lib",
			lib:  "whisper",
			expected: map[string]string{
				"linux":   "/usr/local/lib/libwhisper.so",
				"freebsd": "/usr/local/lib/libwhisper.so",
				"darwin":  "/usr/local/lib/libwhisper.dylib",
				"windows": "/usr/local/lib/whisper.dll",
			},
		},
		{
			name: "ggml library",
			path: "/opt/bucky",
			lib:  "ggml",
			expected: map[string]string{
				"linux":   "/opt/bucky/libggml.so",
				"freebsd": "/opt/bucky/libggml.so",
				"darwin":  "/opt/bucky/libggml.dylib",
				"windows": "/opt/bucky/ggml.dll",
			},
		},
		{
			name: "ggml-cpu library",
			path: "/home/user/libs",
			lib:  "ggml-cpu",
			expected: map[string]string{
				"linux":   "/home/user/libs/libggml-cpu.so",
				"freebsd": "/home/user/libs/libggml-cpu.so",
				"darwin":  "/home/user/libs/libggml-cpu.dylib",
				"windows": "/home/user/libs/ggml-cpu.dll",
			},
		},
		{
			name: "empty path",
			path: "",
			lib:  "whisper",
			expected: map[string]string{
				"linux":   "libwhisper.so",
				"freebsd": "libwhisper.so",
				"darwin":  "libwhisper.dylib",
				"windows": "whisper.dll",
			},
		},
		{
			name: "relative path",
			path: "./lib",
			lib:  "whisper",
			expected: map[string]string{
				"linux":   "lib/libwhisper.so",
				"freebsd": "lib/libwhisper.so",
				"darwin":  "lib/libwhisper.dylib",
				"windows": "lib/whisper.dll",
			},
		},
		{
			name: "path with spaces",
			path: "/path/with spaces",
			lib:  "whisper",
			expected: map[string]string{
				"linux":   "/path/with spaces/libwhisper.so",
				"freebsd": "/path/with spaces/libwhisper.so",
				"darwin":  "/path/with spaces/libwhisper.dylib",
				"windows": "/path/with spaces/whisper.dll",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetLibraryFilename(tt.path, tt.lib)

			expected, ok := tt.expected[runtime.GOOS]
			if !ok {
				if result == "" {
					t.Error("expected non-empty result for unsupported OS")
				}
				return
			}

			expectedNorm := filepath.FromSlash(expected)
			if result != expectedNorm {
				t.Errorf("expected '%s', got '%s'", expectedNorm, result)
			}
		})
	}
}

func TestGetLibraryFilename_CurrentOS(t *testing.T) {
	path := "/test/path"
	lib := "testlib"

	result := GetLibraryFilename(path, lib)

	switch runtime.GOOS {
	case "linux", "freebsd":
		expected := filepath.Join(path, "libtestlib.so")
		if result != expected {
			t.Errorf("expected '%s', got '%s'", expected, result)
		}
	case "darwin":
		expected := filepath.Join(path, "libtestlib.dylib")
		if result != expected {
			t.Errorf("expected '%s', got '%s'", expected, result)
		}
	case "windows":
		expected := filepath.Join(path, "testlib.dll")
		if result != expected {
			t.Errorf("expected '%s', got '%s'", expected, result)
		}
	}
}

func TestGetLibraryFilename_DifferentLibNames(t *testing.T) {
	libs := []string{"whisper", "ggml", "ggml-base", "ggml-cpu"}
	basePath := "/lib"

	for _, lib := range libs {
		t.Run(lib, func(t *testing.T) {
			result := GetLibraryFilename(basePath, lib)

			if result == "" {
				t.Errorf("expected non-empty result for lib '%s'", lib)
			}

			expectedPrefix := filepath.FromSlash(basePath)
			if len(result) < len(expectedPrefix) || result[:len(expectedPrefix)] != expectedPrefix {
				t.Errorf("expected path to start with '%s', got '%s'", expectedPrefix, result)
			}

			if !strings.Contains(result, lib) {
				t.Errorf("expected result to contain '%s', got '%s'", lib, result)
			}
		})
	}
}

// TestLoadLibrary_MissingPath verifies LoadLibrary returns a useful error
// (and does not panic) when neither path nor BUCKY_LIB is supplied.
func TestLoadLibrary_MissingPath(t *testing.T) {
	t.Setenv("BUCKY_LIB", "")

	_, err := LoadLibrary("", "whisper")
	if err == nil {
		t.Fatal("LoadLibrary(\"\", ...): expected error, got nil")
	}
	if !strings.Contains(err.Error(), "BUCKY_LIB") {
		t.Errorf("error message %q should mention BUCKY_LIB", err)
	}
}

// TestLoadLibrary_BadPath verifies LoadLibrary returns an error (and does
// not panic) when the path is set but no library exists at the expected
// filename. dlopen / LoadLibrary on every supported OS reports a clear
// failure here; we just need to confirm the wrapper surfaces it.
func TestLoadLibrary_BadPath(t *testing.T) {
	tmp := t.TempDir()

	_, err := LoadLibrary(tmp, "this-library-does-not-exist")
	if err == nil {
		t.Fatalf("LoadLibrary(%q, ...): expected error, got nil", tmp)
	}
}

// TestLoadLibrary_Success verifies LoadLibrary returns a usable handle for
// the real libwhisper when BUCKY_LIB is set. Skipped otherwise so this
// file passes in environments without the C library installed.
func TestLoadLibrary_Success(t *testing.T) {
	libPath := os.Getenv("BUCKY_LIB")
	if libPath == "" {
		t.Skip("BUCKY_LIB not set; skipping LoadLibrary success test")
	}

	if _, err := LoadLibrary(libPath, "whisper"); err != nil {
		t.Fatalf("LoadLibrary(%q, whisper): %v", libPath, err)
	}
}
