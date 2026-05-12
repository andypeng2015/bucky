package whisper

import (
	"fmt"

	"github.com/ardanlabs/bucky/pkg/loader"
)

var libPath string

// LibPath returns the path to the loaded whisper.cpp shared library.
func LibPath() string {
	return libPath
}

// Load loads the shared whisper.cpp library from the specified path and
// resolves all FFI function pointers used by this package.
//
// The whisper.cpp xcframework on darwin bundles ggml inside libwhisper, so a
// single library load is sufficient. On windows, ggml-base/ggml-cpu/ggml DLLs
// live next to whisper.dll and the OS resolves them automatically when
// libwhisper is loaded.
func Load(path string) error {
	libPath = path

	lib, err := loader.LoadLibrary(path, "whisper")
	if err != nil {
		return err
	}

	if err := loadSystemFuncs(lib); err != nil {
		return err
	}

	if err := loadContextFuncs(lib); err != nil {
		return err
	}

	if err := loadModelFuncs(lib); err != nil {
		return err
	}

	if err := loadParamsFuncs(lib); err != nil {
		return err
	}

	if err := loadFullFuncs(lib); err != nil {
		return err
	}

	if err := loadSegmentsFuncs(lib); err != nil {
		return err
	}

	if err := loadLangFuncs(lib); err != nil {
		return err
	}

	return nil
}

func loadError(name string, err error) error {
	return fmt.Errorf("could not load %q: %w", name, err)
}
