package download

import (
	"archive/zip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	getter "github.com/hashicorp/go-getter"
)

var (
	ErrUnknownArch         = errors.New("unknown architecture")
	ErrUnknownOS           = errors.New("unknown OS")
	ErrUnknownProcessor    = errors.New("unknown processor")
	ErrInvalidVersion      = errors.New("invalid version")
	ErrFileNotFound        = errors.New("could not download file: the requested whisper.cpp version may still be building for your platform")
	ErrLinuxNoPrebuilt     = errors.New("whisper.cpp does not publish prebuilt Linux binaries; build from source per INSTALL.md")
	ErrUnsupportedPlatform = errors.New("no prebuilt whisper.cpp asset for this platform")
)

// DefaultWhisperVersion is the well-known whisper.cpp release tag bucky's
// FFI struct mirrors (e.g. WhisperFullParams's 304-byte layout) are tested
// against. `bucky install` uses this when no -v flag is supplied so first
// installs and CI runs do not depend on the GitHub releases API. Bumping
// this value is a deliberate, reviewable change that should be paired with
// re-running the FFI sizeof + by-ref/by-value tests in pkg/whisper.
const DefaultWhisperVersion = "v1.8.4"

var (
	// RetryCount is how many times the package will retry to obtain the latest whisper.cpp version.
	RetryCount = 3
	// RetryDelay is the delay between retries when obtaining the latest whisper.cpp version.
	RetryDelay = 3 * time.Second
	// versionURL is the URL for fetching the latest whisper.cpp release.
	versionURL = "https://api.github.com/repos/ggml-org/whisper.cpp/releases/latest"
)

// WhisperLatestVersion fetches the latest release tag of whisper.cpp from the
// upstream GitHub releases API.
func WhisperLatestVersion() (string, error) {
	var version string
	var err error
	for range RetryCount {
		version, err = getLatestVersion()
		if err == nil {
			return version, nil
		}
		time.Sleep(RetryDelay)
	}

	return "", errors.New("unable to fetch latest version")
}

func getLatestVersion() (string, error) {
	req, err := http.NewRequest("GET", versionURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("received status code %d from version URL: %s", resp.StatusCode, string(body))
	}

	var result struct {
		TagName string `json:"tag_name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.TagName, nil
}

// getDownloadLocationAndFilename returns the download URL location and the
// asset filename for the given parameters.
func getDownloadLocationAndFilename(arch Arch, os OS, prcssr Processor, version string) (location, filename string, err error) {
	location = fmt.Sprintf("https://github.com/ggml-org/whisper.cpp/releases/download/%s", version)

	switch os {
	case Darwin:
		// The xcframework is universal (arm64 + x86_64) and includes Metal.
		switch prcssr {
		case CPU, Metal:
			filename = fmt.Sprintf("whisper-%s-xcframework.zip", version)
		default:
			return "", "", fmt.Errorf("%w: darwin only supports cpu/metal", ErrUnknownProcessor)
		}

	case Windows:
		if arch != AMD64 {
			return "", "", fmt.Errorf("%w: windows %s not supported in v1", ErrUnsupportedPlatform, arch)
		}
		switch prcssr {
		case CPU:
			filename = "whisper-bin-x64.zip"
		case CUDA:
			filename = "whisper-cublas-12.4.0-bin-x64.zip"
		default:
			return "", "", fmt.Errorf("%w: windows supports cpu/cuda", ErrUnknownProcessor)
		}

	case Linux:
		return "", "", ErrLinuxNoPrebuilt

	default:
		return "", "", ErrUnknownOS
	}

	return location, filename, nil
}

// getFunc is the function used to download asset files. It can be overridden for testing.
var getFunc = get

// Get downloads the whisper.cpp precompiled binaries for the desired arch/OS/processor.
//
//	arch:      "amd64" or "arm64"
//	os:        "linux", "darwin", or "windows"
//	processor: "cpu", "cuda", "metal", or "vulkan"
//	version:   the desired whisper.cpp release tag, e.g. "v1.8.4"
//	dest:      destination directory for the extracted libraries
func Get(architecture string, operatingSystem string, processor string, version string, dest string) error {
	return GetWithProgress(architecture, operatingSystem, processor, version, dest, ProgressTracker)
}

// GetWithProgress downloads the whisper.cpp precompiled binaries using the
// provided progress tracker.
func GetWithProgress(architecture string, operatingSystem string, processor string, version string, dest string, progress getter.ProgressTracker) error {
	return GetWithContext(context.Background(), architecture, operatingSystem, processor, version, dest, progress)
}

// GetWithContext downloads the whisper.cpp precompiled binaries using the
// provided context and progress tracker.
func GetWithContext(ctx context.Context, architecture string, operatingSystem string, processor string, version string, dest string, progress getter.ProgressTracker) error {
	arch, err := ParseArch(architecture)
	if err != nil {
		return ErrUnknownArch
	}

	osVal, err := ParseOS(operatingSystem)
	if err != nil {
		return ErrUnknownOS
	}

	prcssr, err := ParseProcessor(processor)
	if err != nil {
		return ErrUnknownProcessor
	}

	if err := VersionIsValid(version); err != nil {
		return ErrInvalidVersion
	}

	location, filename, err := getDownloadLocationAndFilename(arch, osVal, prcssr, version)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/%s", location, filename)
	return getFunc(ctx, url, dest, osVal, progress)
}

// get downloads the asset zip and extracts the relevant whisper library file(s)
// into dest. The extraction logic differs per OS because whisper.cpp ships
// platform-specific archive layouts.
func get(ctx context.Context, url, dest string, osVal OS, progress getter.ProgressTracker) error {
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return fmt.Errorf("failed to create destination dir: %w", err)
	}

	downloadFile := filepath.Join(dest, filepath.Base(url))

	client := &getter.Client{
		Ctx:  ctx,
		Src:  url + "?archive=false",
		Dst:  dest,
		Mode: getter.ClientModeAny,
	}

	if progress != nil {
		client.ProgressListener = progress
	}

	if err := client.Get(); err != nil {
		if strings.Contains(err.Error(), "404") {
			return fmt.Errorf("%w: %s", ErrFileNotFound, url)
		}
		return err
	}
	defer os.Remove(downloadFile)

	switch osVal {
	case Darwin:
		return extractDarwinXCFramework(downloadFile, dest)
	case Windows:
		return extractWindowsZip(downloadFile, dest)
	default:
		return fmt.Errorf("%w: extraction not implemented for %s", ErrUnsupportedPlatform, osVal)
	}
}

// extractDarwinXCFramework pulls the macos-arm64_x86_64 universal dylib out
// of the xcframework zip and writes it to dest as libwhisper.dylib.
func extractDarwinXCFramework(zipPath, dest string) error {
	const wantPath = "build-apple/whisper.xcframework/macos-arm64_x86_64/whisper.framework/Versions/A/whisper"

	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open xcframework zip: %w", err)
	}
	defer zr.Close()

	for _, f := range zr.File {
		if f.Name != wantPath {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("failed to open dylib in zip: %w", err)
		}
		out, err := os.OpenFile(filepath.Join(dest, "libwhisper.dylib"), os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o755)
		if err != nil {
			rc.Close()
			return fmt.Errorf("failed to create libwhisper.dylib: %w", err)
		}
		if _, err := io.Copy(out, rc); err != nil {
			rc.Close()
			out.Close()
			return fmt.Errorf("failed to write libwhisper.dylib: %w", err)
		}
		rc.Close()
		out.Close()
		return nil
	}

	return fmt.Errorf("xcframework zip did not contain %s", wantPath)
}

// extractWindowsZip pulls all DLLs out of the Release/ directory of the
// windows release zip and writes them to dest.
func extractWindowsZip(zipPath, dest string) error {
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open windows zip: %w", err)
	}
	defer zr.Close()

	any := false
	for _, f := range zr.File {
		// Only extract DLLs from the Release/ directory.
		base := filepath.Base(f.Name)
		if !strings.HasSuffix(strings.ToLower(base), ".dll") {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("failed to open %s in zip: %w", f.Name, err)
		}
		target := filepath.Join(dest, base)
		out, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o755)
		if err != nil {
			rc.Close()
			return fmt.Errorf("failed to create %s: %w", target, err)
		}
		if _, err := io.Copy(out, rc); err != nil {
			rc.Close()
			out.Close()
			return fmt.Errorf("failed to write %s: %w", target, err)
		}
		rc.Close()
		out.Close()
		any = true
	}

	if !any {
		return errors.New("windows zip contained no DLL files")
	}

	return nil
}

// VersionIsValid checks if the provided version string looks like a
// whisper.cpp release tag (e.g. "v1.8.4").
func VersionIsValid(version string) error {
	if !strings.HasPrefix(version, "v") {
		return ErrInvalidVersion
	}
	// Cheap shape check: must contain at least one dot.
	if !strings.Contains(version, ".") {
		return ErrInvalidVersion
	}
	return nil
}

// LibraryName returns the filename for the whisper.cpp library on the given OS.
func LibraryName(operatingSystem string) string {
	osVal, err := ParseOS(operatingSystem)
	if err != nil {
		return "unknown"
	}

	switch osVal {
	case Linux:
		return "libwhisper.so"
	case Windows:
		return "whisper.dll"
	case Darwin:
		return "libwhisper.dylib"
	default:
		return "unknown"
	}
}
