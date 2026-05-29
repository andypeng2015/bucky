package whisper

import (
	"sync"
	"testing"
)

// TestVersion verifies the whisper.cpp version string crosses the FFI
// boundary as a non-empty Go string. An empty result would indicate the
// return-value marshalling of the void→const-char* trampoline broke.
func TestVersion(t *testing.T) {
	testSetup(t)

	v := Version()
	if v == "" {
		t.Fatal("Version returned empty string")
	}
	t.Logf("whisper.Version = %q", v)
}

// TestPrintSystemInfo verifies the system-info string is returned intact.
func TestPrintSystemInfo(t *testing.T) {
	testSetup(t)

	info := PrintSystemInfo()
	if info == "" {
		t.Fatal("PrintSystemInfo returned empty string")
	}
	t.Logf("whisper.PrintSystemInfo = %q", info)
}

// TestContextConcurrent stands up several independent whisper Context
// handles on the same model file and exercises read-only model accessors
// against each handle from its own goroutine. Production callers in bucky
// (and downstream consumers) routinely allocate one Context per inflight
// transcription job and share the loaded library across goroutines, so the
// real concurrency contract is "many Contexts, one library" — not "many
// goroutines hammering one global accessor".
//
// The functional assertions are intentionally light. The point of the test
// is to give `go test -race` a real workload to inspect: the per-handle
// trampoline calls, the shared ffi.Fun values, and the upstream whisper
// model-state reads all run side-by-side. A race or a torn struct here is
// a real bug, not a theoretical one.
//
// Requires BUCKY_LIB and BUCKY_TEST_MODEL; skipped otherwise. Kept small
// (4 instances) because each InitFromFileWithParams call mmaps the model
// and the smallest tiny.en weights are still ~75 MB.
func TestContextConcurrent(t *testing.T) {
	testSetup(t)
	modelPath := testModelFileName(t)

	const instances = 4
	const iterations = 25

	ctxs := make([]Context, instances)
	for i := range ctxs {
		cparams := testContextDefaultParams(t)
		ctx, err := InitFromFileWithParams(modelPath, cparams)
		if err != nil {
			t.Fatalf("InitFromFileWithParams[%d]: %v", i, err)
		}
		ctxs[i] = ctx
	}
	defer func() {
		for _, ctx := range ctxs {
			Free(ctx)
		}
	}()

	// Snapshot per-instance expected values once, single-threaded, so the
	// fan-out assertions have a known-good baseline.
	wantVocab := ModelNVocab(ctxs[0])
	wantAudioCtx := ModelNAudioCtx(ctxs[0])
	wantTextCtx := ModelNTextCtx(ctxs[0])
	wantMels := ModelNMels(ctxs[0])
	wantType := ModelTypeReadable(ctxs[0])
	if wantVocab <= 0 || wantAudioCtx <= 0 || wantTextCtx <= 0 || wantMels <= 0 || wantType == "" {
		t.Fatalf("baseline accessors returned zero values: vocab=%d audio=%d text=%d mels=%d type=%q",
			wantVocab, wantAudioCtx, wantTextCtx, wantMels, wantType)
	}

	var wg sync.WaitGroup
	wg.Add(instances)
	for i, ctx := range ctxs {
		go func() {
			defer wg.Done()
			for range iterations {
				if v := ModelNVocab(ctx); v != wantVocab {
					t.Errorf("instance %d: ModelNVocab = %d, want %d", i, v, wantVocab)
					return
				}
				if v := ModelNAudioCtx(ctx); v != wantAudioCtx {
					t.Errorf("instance %d: ModelNAudioCtx = %d, want %d", i, v, wantAudioCtx)
					return
				}
				if v := ModelNTextCtx(ctx); v != wantTextCtx {
					t.Errorf("instance %d: ModelNTextCtx = %d, want %d", i, v, wantTextCtx)
					return
				}
				if v := ModelNMels(ctx); v != wantMels {
					t.Errorf("instance %d: ModelNMels = %d, want %d", i, v, wantMels)
					return
				}
				if v := ModelTypeReadable(ctx); v != wantType {
					t.Errorf("instance %d: ModelTypeReadable = %q, want %q", i, v, wantType)
					return
				}
			}
		}()
	}
	wg.Wait()
}
