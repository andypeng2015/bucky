package whisper

import (
	"unsafe"

	"github.com/ardanlabs/bucky/pkg/utils"
	"github.com/jupiterrider/ffi"
)

var (
	// WHISPER_API const char * whisper_version(void);
	versionFunc ffi.Fun

	// WHISPER_API const char * whisper_print_system_info(void);
	printSystemInfoFunc ffi.Fun

	// GGML_API void ggml_backend_load_all_from_path(const char * dir_path);
	//
	// Re-exported transitively via libwhisper.so → libggml-base.so on
	// Linux. May be missing on macOS xcframework builds (where backends
	// are statically linked); the loader treats a missing symbol as a
	// soft no-op.
	ggmlBackendLoadAllFromPathFunc ffi.Fun
)

func loadSystemFuncs(lib ffi.Lib) error {
	var err error

	if versionFunc, err = lib.Prep("whisper_version", &ffi.TypePointer); err != nil {
		return loadError("whisper_version", err)
	}

	if printSystemInfoFunc, err = lib.Prep("whisper_print_system_info", &ffi.TypePointer); err != nil {
		return loadError("whisper_print_system_info", err)
	}

	// Optional: only present when libwhisper was built with
	// -DGGML_BACKEND_DL=ON (bucky-builder's Linux artifacts). Best-effort
	// — a Prep failure here just means the symbol isn't exported, which
	// is fine for static builds.
	if fn, perr := lib.Prep("ggml_backend_load_all_from_path", &ffi.TypeVoid, &ffi.TypePointer); perr == nil {
		ggmlBackendLoadAllFromPathFunc = fn
	}

	return nil
}

// Version returns the whisper.cpp library version string.
func Version() string {
	var ptr *byte
	versionFunc.Call(unsafe.Pointer(&ptr))
	if ptr == nil {
		return ""
	}
	return utils.BytePtrToString(ptr)
}

// PrintSystemInfo returns the system info string reported by whisper.cpp
// (the same string the upstream CLI prints at startup).
func PrintSystemInfo() string {
	var ptr *byte
	printSystemInfoFunc.Call(unsafe.Pointer(&ptr))
	if ptr == nil {
		return ""
	}
	return utils.BytePtrToString(ptr)
}

// ggmlBackendLoadAllFromPath dlopens every libggml-*.{so,dylib,dll} found in
// dirPath so each backend's static constructor self-registers with ggml.
//
// Required for builds where ggml backends ship as separate dynamic
// libraries (-DGGML_BACKEND_DL=ON), which is how bucky-builder produces the
// Linux artifacts. Without this call, ggml ends up with zero registered
// backends and the first model init asserts on device==NULL.
//
// On builds where backends are statically linked into libwhisper (the
// upstream macOS xcframework, the upstream Windows zip), the symbol is not
// present at all and this is a no-op (loadSystemFuncs leaves the Fun
// zero-valued). On builds where the symbol is present but no
// libggml-*.{so,dylib,dll} files match in dirPath, the underlying ggml
// implementation is itself a safe no-op.
func ggmlBackendLoadAllFromPath(dirPath string) error {
	if ggmlBackendLoadAllFromPathFunc == (ffi.Fun{}) {
		return nil
	}
	cpath, err := utils.BytePtrFromString(dirPath)
	if err != nil {
		return err
	}
	ggmlBackendLoadAllFromPathFunc.Call(nil, unsafe.Pointer(&cpath))
	return nil
}
