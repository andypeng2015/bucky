package whisper

import (
	"os"
	"sync"
	"testing"
)

var loadOnce sync.Once
var loadErr error

func testSetup(t *testing.T) {
	t.Helper()

	libPath := os.Getenv("BUCKY_LIB")
	if libPath == "" {
		t.Skip("BUCKY_LIB not set; skipping whisper FFI test")
	}

	loadOnce.Do(func() {
		loadErr = Load(libPath)
	})
	if loadErr != nil {
		t.Fatalf("failed to load whisper.cpp from %s: %v", libPath, loadErr)
	}
}

func testModelFileName(t *testing.T) string {
	t.Helper()
	model := os.Getenv("BUCKY_TEST_MODEL")
	if model == "" {
		t.Skip("BUCKY_TEST_MODEL not set; skipping test that requires a model")
	}
	if _, err := os.Stat(model); err != nil {
		t.Skipf("model file %q not present: %v", model, err)
	}
	return model
}

func testAudioFileName(t *testing.T) string {
	t.Helper()
	audio := os.Getenv("BUCKY_TEST_AUDIO")
	if audio == "" {
		t.Skip("BUCKY_TEST_AUDIO not set; skipping test that requires an audio sample")
	}
	if _, err := os.Stat(audio); err != nil {
		t.Skipf("audio file %q not present: %v", audio, err)
	}
	return audio
}

// testContextDefaultParams returns the C library's default context params
// with one CI-friendly tweak: when BUCKY_USE_GPU=0 is set in the environment
// it forces UseGPU off. The bucky-builder CPU-only Linux artifacts still
// have GPU init scaffolding compiled in but no GPU backend registered, so
// the upstream default (use_gpu=1) makes ggml_backend_dev_init assert with
// device==NULL. macOS and the cuda/vulkan Linux artifacts have a real GPU
// backend so the env var is unset there and use_gpu stays at its default.
func testContextDefaultParams(t *testing.T) ContextParams {
	t.Helper()
	p := ContextDefaultParams()
	if os.Getenv("BUCKY_USE_GPU") == "0" {
		p.UseGPU = 0
	}
	return p
}
