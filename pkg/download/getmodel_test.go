package download

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

// TestGetModelSuccess drives getModel against an in-process server and
// verifies the bytes land on disk intact.
//
// getModel is called in production from cmd/model.go with `dest` set to
// the parent directory (e.g. ~/models), and go-getter writes the URL's
// basename into that directory under ClientModeAny. The test mirrors
// that contract: pass a temp dir as dest, then read the basename back.
func TestGetModelSuccess(t *testing.T) {
	payload := []byte("hello whisper")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "13")
		_, _ = w.Write(payload)
	}))
	defer srv.Close()

	destDir := t.TempDir()
	if err := getModel(context.Background(), srv.URL+"/ggml-fake.bin", destDir, nil); err != nil {
		t.Fatalf("getModel: %v", err)
	}

	target := filepath.Join(destDir, "ggml-fake.bin")
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !reflect.DeepEqual(got, payload) {
		t.Errorf("payload: got %q, want %q", got, payload)
	}
}

// TestGetModelContextCanceled drives getModel against a server that blocks
// forever, cancels the context mid-flight, and verifies the call aborts.
// A bug that severs ctx propagation through go-getter would block this
// test until the test framework's deadline expires; that is the
// regression we want to fast-fail.
func TestGetModelContextCanceled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	destDir := t.TempDir()

	done := make(chan error, 1)
	go func() { done <- getModel(ctx, srv.URL+"/slow.bin", destDir, nil) }()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("getModel: expected error after ctx cancel, got nil")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("getModel: did not return within 5s after ctx cancel")
	}
}

// TestGetModelHTTPError verifies an upstream 404 surfaces as an error
// rather than silently writing an HTML body to the target path.
func TestGetModelHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	destDir := t.TempDir()
	if err := getModel(context.Background(), srv.URL+"/missing.bin", destDir, nil); err == nil {
		t.Fatal("getModel: expected error on 404, got nil")
	}
}
