package whisper

import (
	"unsafe"

	"github.com/ardanlabs/bucky/pkg/utils"
	"github.com/jupiterrider/ffi"
)

var (
	// WHISPER_API int whisper_full_n_segments(struct whisper_context * ctx);
	fullNSegmentsFunc ffi.Fun

	// WHISPER_API const char * whisper_full_get_segment_text(struct whisper_context * ctx, int i_segment);
	fullGetSegmentTextFunc ffi.Fun

	// WHISPER_API int64_t whisper_full_get_segment_t0(struct whisper_context * ctx, int i_segment);
	fullGetSegmentT0Func ffi.Fun

	// WHISPER_API int64_t whisper_full_get_segment_t1(struct whisper_context * ctx, int i_segment);
	fullGetSegmentT1Func ffi.Fun
)

func loadSegmentsFuncs(lib ffi.Lib) error {
	var err error

	if fullNSegmentsFunc, err = lib.Prep("whisper_full_n_segments", &ffi.TypeSint32, &ffi.TypePointer); err != nil {
		return loadError("whisper_full_n_segments", err)
	}

	if fullGetSegmentTextFunc, err = lib.Prep("whisper_full_get_segment_text", &ffi.TypePointer, &ffi.TypePointer, &ffi.TypeSint32); err != nil {
		return loadError("whisper_full_get_segment_text", err)
	}

	if fullGetSegmentT0Func, err = lib.Prep("whisper_full_get_segment_t0", &ffi.TypeSint64, &ffi.TypePointer, &ffi.TypeSint32); err != nil {
		return loadError("whisper_full_get_segment_t0", err)
	}

	if fullGetSegmentT1Func, err = lib.Prep("whisper_full_get_segment_t1", &ffi.TypeSint64, &ffi.TypePointer, &ffi.TypeSint32); err != nil {
		return loadError("whisper_full_get_segment_t1", err)
	}

	return nil
}

// FullNSegments returns the number of generated text segments.
func FullNSegments(ctx Context) int32 {
	if ctx == 0 {
		return 0
	}
	var result ffi.Arg
	fullNSegmentsFunc.Call(unsafe.Pointer(&result), unsafe.Pointer(&ctx))
	return int32(result)
}

// FullGetSegmentText returns the transcribed text for the given segment index.
func FullGetSegmentText(ctx Context, iSegment int32) string {
	if ctx == 0 {
		return ""
	}
	var ptr *byte
	fullGetSegmentTextFunc.Call(unsafe.Pointer(&ptr), unsafe.Pointer(&ctx), unsafe.Pointer(&iSegment))
	if ptr == nil {
		return ""
	}
	return utils.BytePtrToString(ptr)
}

// FullGetSegmentT0 returns the start time of the segment in 10 ms units.
// Multiply by 10 to get milliseconds.
func FullGetSegmentT0(ctx Context, iSegment int32) int64 {
	if ctx == 0 {
		return 0
	}
	var result int64
	fullGetSegmentT0Func.Call(unsafe.Pointer(&result), unsafe.Pointer(&ctx), unsafe.Pointer(&iSegment))
	return result
}

// FullGetSegmentT1 returns the end time of the segment in 10 ms units.
// Multiply by 10 to get milliseconds.
func FullGetSegmentT1(ctx Context, iSegment int32) int64 {
	if ctx == 0 {
		return 0
	}
	var result int64
	fullGetSegmentT1Func.Call(unsafe.Pointer(&result), unsafe.Pointer(&ctx), unsafe.Pointer(&iSegment))
	return result
}
