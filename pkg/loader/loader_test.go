package loader

import (
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
