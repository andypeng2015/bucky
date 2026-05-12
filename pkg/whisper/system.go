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
)

func loadSystemFuncs(lib ffi.Lib) error {
	var err error

	if versionFunc, err = lib.Prep("whisper_version", &ffi.TypePointer); err != nil {
		return loadError("whisper_version", err)
	}

	if printSystemInfoFunc, err = lib.Prep("whisper_print_system_info", &ffi.TypePointer); err != nil {
		return loadError("whisper_print_system_info", err)
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
