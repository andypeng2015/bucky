package whisper

import (
	"unsafe"

	"github.com/ardanlabs/bucky/pkg/utils"
	"github.com/jupiterrider/ffi"
)

var (
	// WHISPER_API int whisper_lang_max_id(void);
	langMaxIdFunc ffi.Fun

	// WHISPER_API int whisper_lang_id(const char * lang);
	langIdFunc ffi.Fun

	// WHISPER_API const char * whisper_lang_str(int id);
	langStrFunc ffi.Fun
)

func loadLangFuncs(lib ffi.Lib) error {
	var err error

	if langMaxIdFunc, err = lib.Prep("whisper_lang_max_id", &ffi.TypeSint32); err != nil {
		return loadError("whisper_lang_max_id", err)
	}

	if langIdFunc, err = lib.Prep("whisper_lang_id", &ffi.TypeSint32, &ffi.TypePointer); err != nil {
		return loadError("whisper_lang_id", err)
	}

	if langStrFunc, err = lib.Prep("whisper_lang_str", &ffi.TypePointer, &ffi.TypeSint32); err != nil {
		return loadError("whisper_lang_str", err)
	}

	return nil
}

// LangMaxId returns the largest language id (i.e. number of languages - 1).
func LangMaxId() int32 {
	var result ffi.Arg
	langMaxIdFunc.Call(unsafe.Pointer(&result))
	return int32(result)
}

// LangId returns the id of the specified language code (e.g. "de" -> 2),
// or -1 if the language is unknown.
func LangId(lang string) int32 {
	cstr, err := utils.BytePtrFromString(lang)
	if err != nil {
		return -1
	}
	var result ffi.Arg
	langIdFunc.Call(unsafe.Pointer(&result), unsafe.Pointer(&cstr))
	return int32(result)
}

// LangStr returns the short language code for the given id (e.g. 2 -> "de"),
// or an empty string if the id is invalid.
func LangStr(id int32) string {
	var ptr *byte
	langStrFunc.Call(unsafe.Pointer(&ptr), unsafe.Pointer(&id))
	if ptr == nil {
		return ""
	}
	return utils.BytePtrToString(ptr)
}
