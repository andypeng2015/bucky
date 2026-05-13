package whisper

import (
	"runtime"
	"unsafe"

	"github.com/ardanlabs/bucky/pkg/utils"
)

// StringRefs collects C-string pointers that back string fields in a
// WhisperFullParams. The caller must keep the StringRefs value alive (e.g.
// via runtime.KeepAlive) for as long as the params struct is in use, so the
// underlying byte buffers are not garbage-collected.
type StringRefs struct {
	keep []*byte
}

// SetLanguage sets params.Language from a Go string. Pass "" or "auto" to
// have whisper auto-detect.
func (s *StringRefs) SetLanguage(p *WhisperFullParams, lang string) error {
	if lang == "" {
		p.Language = 0
		return nil
	}
	b, err := utils.BytePtrFromString(lang)
	if err != nil {
		return err
	}
	s.keep = append(s.keep, b)
	p.Language = uintptr(unsafe.Pointer(b))
	return nil
}

// SetInitialPrompt sets params.InitialPrompt from a Go string.
func (s *StringRefs) SetInitialPrompt(p *WhisperFullParams, prompt string) error {
	if prompt == "" {
		p.InitialPrompt = 0
		return nil
	}
	b, err := utils.BytePtrFromString(prompt)
	if err != nil {
		return err
	}
	s.keep = append(s.keep, b)
	p.InitialPrompt = uintptr(unsafe.Pointer(b))
	return nil
}

// SetSuppressRegex sets params.SuppressRegex from a Go string.
func (s *StringRefs) SetSuppressRegex(p *WhisperFullParams, re string) error {
	if re == "" {
		p.SuppressRegex = 0
		return nil
	}
	b, err := utils.BytePtrFromString(re)
	if err != nil {
		return err
	}
	s.keep = append(s.keep, b)
	p.SuppressRegex = uintptr(unsafe.Pointer(b))
	return nil
}

// KeepAlive marks all backing buffers reachable. Defer this call right after
// constructing your StringRefs so the strings outlive the FFI call.
func (s *StringRefs) KeepAlive() {
	runtime.KeepAlive(s.keep)
}
