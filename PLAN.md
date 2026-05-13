# Bucky — Project Plan

`bucky` is a Go FFI binding to [whisper.cpp](https://github.com/ggml-org/whisper.cpp),
modeled directly on [hybridgroup/yzma](https://github.com/hybridgroup/yzma) (which
binds llama.cpp/mtmd). The end goal is to give Kronk a native, high-quality
speech-to-text path so we can implement OpenAI-compatible
`POST /v1/audio/transcriptions` (kronk issue #565) without prompt-hacking an
OMNI chat model.

> **Naming**: Bucky is the squirrel from *The Emperor's New Groove*. He
> communicates in squeaks/whispers — perfect mascot for a speech-to-text
> binding. Kronk and Yzma are the existing siblings; future TTS work could
> naturally be `kuzco`.

---

## Status at a glance

| PR | Scope | Status |
|---|---|---|
| #1 | Foundation: CLI shell, `pkg/loader`, `pkg/download`, `bucky install` | ✅ landed |
| #2 | Core FFI binding: `pkg/whisper` for model + `whisper_full` + segments | ✅ landed |
| #3 | Audio decoding (`pkg/audio`), per-token API, `bucky whisper transcribe`, examples, docs | ✅ landed |
| #4 | VAD + state + word-level timestamps + bench helpers + streaming example | ✅ landed |
| #5 | Kronk integration: `transcribeapp/`, `WhisperPool`, OpenAI-compatible HTTP endpoint | ⏳ next |

**What's left, in priority order**

1. **Windows smoke run**: `pkg/whisper` is exercised on `windows-latest`
   by GitHub Actions for build/vet/staticcheck, but the FFI sizeof +
   by-ref/by-value round-trip tests still need a Windows host with
   `whisper.dll`. See INSTALL.md "Verification gap".
2. **PR #5** (the original goal): wire bucky into the kronk repo,
   implement `POST /v1/audio/transcriptions` (kronk issue #565).

---

## How to resume this work in a new session

Tell Dave (the agent):

> "Continue work on bucky. Read `/Users/bill/code/go/src/github.com/ardanlabs/bucky/PLAN.md`
> and pick up at PR #N."

Required context every session:

- **bucky repo**: `/Users/bill/code/go/src/github.com/ardanlabs/bucky`
- **yzma reference repo (read-only model)**: `/Users/bill/code/go/src/github.com/hybridgroup/yzma`
- **kronk repo (downstream consumer)**: `/Users/bill/code/go/src/github.com/ardanlabs/kronk`
- **kronk issue driving this**: <https://github.com/ardanlabs/kronk/issues/565>
- **upstream whisper.cpp**: <https://github.com/ggml-org/whisper.cpp>

---

## Decisions already made

| # | Decision | Rationale |
|---|---|---|
| 1 | Build a separate repo `github.com/ardanlabs/bucky` instead of extending Kronk or yzma | Whisper is a separate runtime from llama.cpp and deserves its own clean lifecycle |
| 2 | Mirror yzma's layout file-for-file where applicable | Drop-in mental model for anyone who knows yzma; easy review |
| 3 | License: **Apache-2.0** (as already created in repo) | Patent grant + retaliation clause; corporate-friendly. Will not switch to BSD-3 |
| 4 | FFI strategy: purego + jupiterrider/ffi (same as yzma) — **no CGo** | Consistency with yzma; cross-compile friendly |
| 5 | Audio decoder strategy for v1: **pure-Go decoders** (wav, mp3, flac) | Simple, no native deps; covers ~90% of real-world Whisper inputs. Document the m4a/ogg/webm gap; revisit later with dr_libs/miniaudio if needed |
| 6 | Whisper integrates into Kronk **only after** bucky's API is stable | Don't couple them prematurely |
| 7 | Whisper.cpp version pin: **v1.8.4** (released 2026-03-19) | Default in `pkg/download/`; upgradeable via `bucky install -v vX.Y.Z` |
| 8 | Module path: `github.com/ardanlabs/bucky/pkg/whisper` (analog of `yzma/pkg/llama`) | Confirmed |
| 9 | Linux: **ship no prebuilt binary**; `bucky install` returns a clear error pointing at INSTALL.md build-from-source | Upstream has no Linux release artifact |

### Whisper.cpp upstream artifact matrix (v1.8.4, verified)

| Asset | Platform / Arch | Backend | Contains | Bucky strategy |
|---|---|---|---|---|
| `whisper-v1.8.4-xcframework.zip` | darwin arm64 + amd64 (universal) | CPU + Metal | `build-apple/whisper.xcframework/macos-arm64_x86_64/whisper.framework/Versions/A/whisper` (fat Mach-O dylib) | Extract, rename to `libwhisper.dylib` in `lib/` |
| `whisper-bin-x64.zip` | windows amd64 | CPU | `Release/whisper.dll` + `Release/ggml-{base,cpu}.dll` + `Release/ggml.dll` | Extract all DLLs to `lib/` |
| `whisper-cublas-12.4.0-bin-x64.zip` | windows amd64 | CUDA 12 | same + cudart | Extract for `-p cuda` |
| (none) | linux amd64/arm64 | any | — | `bucky install` returns clear error referencing `INSTALL.md` build-from-source |

The xcframework is universal so darwin amd64 and arm64 share the same dylib.
The install code only extracts the `macos-arm64_x86_64` slice.

---

## PR #5 — Kronk integration (separate repo: kronk)

- [ ] Add `bucky` to kronk `go.mod`
- [ ] Extend `sdk/tools/libs` to also fetch whisper.cpp (or wire to
      `bucky install` programmatically)
- [ ] New `cmd/server/app/domain/transcribeapp/` mirroring `chatapp/`
      and `embedapp/`
- [ ] New SDK type: separate `WhisperPool` (do NOT merge into `Kronk`
      facade — Whisper is a different runtime)
- [ ] Wire JWT auth, rate limiter, Prometheus metrics, traces
- [ ] Multipart upload handler with configurable max size (default 25 MiB)
- [ ] Response formats v1: `json`, `text`. Reject `verbose_json`,
      `srt`, `vtt` with clear 400 until segment-formatting helper lands
- [ ] Document in `chapter-08-api-endpoints.md`
- [ ] Add `examples/transcribe/` to kronk
- [ ] Tests under `cmd/server/app/domain/transcribeapp/` for: happy
      path JSON, plain-text, missing `file`, missing `model`,
      oversized upload, unsupported `response_format`

---

## Open questions / parking lot

- Do we need to ship our own whisper.cpp binary bundles via a Kronk
  `download-bundle` server endpoint (matching how yzma serves
  `bundle.zip`)? Probably yes once Kronk integration lands; defer.
- ROCm support — yzma has it, whisper.cpp has it, but is anyone asking?
- Word-level timestamps via DTW — `examples/words` ships the basic
  `TokenTimestamps` path; DTW (`ContextParams.DtwTokenTimestamps` +
  `DtwAheadsPreset`) is documented in the example comment but not
  exercised. Revisit if a caller needs tighter alignment.
- Real-time / streaming endpoint for kronk — separate kronk issue;
  do not fold into #565.
- m4a/ogg/webm — wait for user demand. Likely path is shipping
  `dr_libs` (header-only C) via purego rather than CGo or ffmpeg.
- `whisper_init_from_buffer_with_params` and the `_no_state` init
  variants — deferred. `pkg/whisper/model.go` carries a
  `// TODO(PR #4 followup)` note; revisit when a downstream caller
  needs in-memory loading or detached state init.

---

## Conventions

- **No CGo anywhere.** purego + jupiterrider/ffi only, like yzma.
- **Every `.go` file edit**: run `gofmt -s -w`, `go vet`, `staticcheck`
  on the changed package.
- **Tests**: gated on `BUCKY_TEST_*` env vars so `go test ./...` works
  without a model present.
- **Public API**: keep the `pkg/whisper` surface a 1-to-1 mirror of
  `whisper.h` where possible. Higher-level ergonomics live in `cmd/`
  or in downstream projects (Kronk).
- **Versioning**: semver on the Go module. Pin whisper.cpp release
  tag in `pkg/download/`; bump in its own PR with a brief changelog.

---

## Reference points (read these when stuck)

- yzma loader pattern: `hybridgroup/yzma/pkg/loader/loader.go`
- yzma download orchestration: `hybridgroup/yzma/pkg/download/install.go`
- yzma llama binding (closest analog to what `pkg/whisper` becomes):
  `hybridgroup/yzma/pkg/llama/{model,context,batch,sampling}.go`
- yzma CLI: `hybridgroup/yzma/cmd/{install,model,llama}.go`
- yzma Makefile (test env-var pattern): `hybridgroup/yzma/Makefile`
- whisper.h public API: `https://github.com/ggml-org/whisper.cpp/blob/master/include/whisper.h`
- OpenAI transcription API spec: <https://platform.openai.com/docs/api-reference/audio/createTranscription>
