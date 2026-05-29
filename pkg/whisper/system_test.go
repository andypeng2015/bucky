package whisper

import (
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

// NOTE: there is intentionally no concurrent-access test for Context.
// Upstream whisper.cpp documents whisper_full as "Not thread safe for
// same context" (see the doc comment on Context in whisper.go). Callers
// that want parallel transcription must allocate one Context per
// goroutine; that pattern is exercised by the existing single-instance
// tests, so a multi-instance test would only burn CI minutes loading N
// copies of the same model without verifying any new contract.
