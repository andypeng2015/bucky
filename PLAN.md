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

## Decisions still open (resolve before / during PR #1)

- **Whisper.cpp version pin**: latest stable as of plan-write is **v1.8.4**
  (released 2026-03-19). Pin via the `bucky install -v vX.Y.Z` flag default
  in `pkg/download/`.
- **Release-artifact matrix** — *investigated, results below*. Linux gap is
  real; for v1 we ship darwin + windows from upstream and document the
  Linux build-from-source path. A bucky-builder repo for Linux bundles
  is deferred until kronk integration demands it.
- **Module path inside repo**: import as `github.com/ardanlabs/bucky/pkg/whisper`
  (analog of `yzma/pkg/llama`) — confirmed unless we discover a reason to differ.

### Whisper.cpp upstream artifact matrix (v1.8.4, verified)

| Asset | Platform / Arch | Backend | Contains | Bucky strategy |
|---|---|---|---|---|
| `whisper-v1.8.4-xcframework.zip` | darwin arm64 + amd64 (universal) | CPU + Metal | `build-apple/whisper.xcframework/macos-arm64_x86_64/whisper.framework/Versions/A/whisper` (fat Mach-O dylib) | Extract, rename to `libwhisper.dylib` in `lib/` |
| `whisper-bin-x64.zip` | windows amd64 | CPU | `Release/whisper.dll` + `Release/ggml-{base,cpu}.dll` + `Release/ggml.dll` | Extract all DLLs to `lib/` |
| `whisper-bin-Win32.zip` | windows 386 | CPU | same layout | (not in v1; document) |
| `whisper-blas-bin-x64.zip` | windows amd64 | BLAS | same layout | (deferred) |
| `whisper-cublas-12.4.0-bin-x64.zip` | windows amd64 | CUDA 12 | same + cudart | Extract for `-p cuda` |
| `whisper-cublas-11.8.0-bin-x64.zip` | windows amd64 | CUDA 11 | same | (deferred — default to 12) |
| (none) | linux amd64/arm64 | any | — | `bucky install` returns clear error referencing `INSTALL.md` build-from-source |

Note: The xcframework also bundles tvOS / iOS / iOS-sim / mac-catalyst
slices we do not need; the install code only extracts the
`macos-arm64_x86_64` slice. The xcframework is universal so darwin amd64
and arm64 share the same dylib.

---

## Target repo layout (mirrors yzma)

```
github.com/ardanlabs/bucky/
├── bucky.go                ─ package doc (mirrors yzma.go)
├── version.go / version_test.go
├── main.go                 ─ CLI entry
├── Makefile                ─ build, test, install, download-whisper.cpp, download-models
├── README.md / INSTALL.md / MODELS.md / ROADMAP.md / BENCHMARKS.md
├── lib/                    ─ default install dir for libwhisper.{so,dylib,dll}
├── cmd/                    ─ CLI subcommands (cobra-style, like yzma's cmd/)
│   ├── info.go
│   ├── install.go          ─ "bucky install" → fetch whisper.cpp release
│   ├── model.go            ─ "bucky model get -u <hf-url>"
│   ├── whisper.go          ─ "bucky whisper transcribe sample.wav"
│   ├── system.go
│   └── README.md
├── examples/
│   ├── hello/              ─ smallest possible transcription
│   ├── installer/          ─ programmatic install
│   ├── transcribe/         ─ language, prompt, temperature
│   ├── translate/          ─ → English
│   ├── segments/           ─ word/segment timing
│   ├── streaming/          ─ chunked usage
│   └── systeminfo/
└── pkg/
    ├── download/           ─ direct analog of yzma/pkg/download
    │   ├── arch.go os.go processor.go
    │   ├── install.go install_test.go
    │   ├── models.go  models_test.go
    │   ├── download.go download_test.go
    │   └── progress.go
    ├── loader/             ─ verbatim copy of yzma/pkg/loader
    │                          (LoadLibrary by short name "whisper")
    ├── audio/              ─ NEW: decode wav/mp3/flac → 16kHz mono f32 PCM
    │   ├── audio.go
    │   ├── wav.go     wav_test.go
    │   ├── mp3.go     mp3_test.go
    │   ├── flac.go    flac_test.go
    │   └── resample.go resample_test.go
    ├── whisper/            ─ analog of yzma/pkg/llama
    │   ├── whisper.go            (types/constants from whisper.h)
    │   ├── context.go context_test.go
    │   ├── model.go    model_test.go
    │   ├── params.go   params_test.go
    │   ├── full.go     full_test.go     (whisper_full / _full_parallel)
    │   ├── segments.go segments_test.go
    │   ├── tokens.go               (per-token data, timestamps, probs)
    │   ├── state.go                (whisper_state / parallel decoders)
    │   ├── lang.go                 (lang_id, lang_str, auto-detect)
    │   ├── vad.go                  (built-in VAD if upstream is stable)
    │   ├── benchmark_test.go
    │   └── testmain_test.go
    └── utils/
```

### What's intentionally different from yzma

- **`pkg/audio`** is new — Whisper requires 16kHz mono f32 PCM input and
  yzma has no equivalent because chat/VLM inputs are tokens or images.
- **`pkg/whisper`** mirrors `pkg/llama` only where the whisper.h API
  has analogs. Whisper has no KV cache / draft / LoRA equivalents.
- **`pkg/loader` is a copy** for now. If a third sibling project ever
  appears, factor out to a shared module.

---

## PR plan

### PR #1 — Foundation  ✅ landed locally; CI workflow + final review pending

Goal: a buildable, lintable, installable shell with `bucky install`
fetching the whisper.cpp library to `lib/`. No FFI yet.

Tasks:

- [x] `go mod init github.com/ardanlabs/bucky`
- [x] `bucky.go` package doc + `version.go` + `version_test.go`
- [x] `Makefile` with: `build`, `install`, `test`, `download-whisper.cpp`,
      `download-models`, `clean-whisper.cpp`
- [x] `pkg/loader/` — port from yzma, `BUCKY_LIB` env var,
      short-name lookup `LoadLibrary(path, "whisper")`
- [x] `pkg/download/` — `arch.go`, `os.go`, `processor.go`, `progress.go`,
      `models.go`, `install.go` (CUDA detect), `download.go`
      (latest-version via GitHub releases, xcframework extraction for
      darwin, DLL extraction for windows, explicit Linux error)
- [x] `cmd/install.go` — `bucky install -lib ./lib [-v <tag>] [-p <proc>] [--os <os>]`
- [x] `cmd/info.go` and `cmd/system.go` (system.go shows host info now;
      whisper FFI hookups land in PR #2)
- [x] `main.go` wiring up the CLI (install, system, version, info)
- [x] `README.md` quickstart, `INSTALL.md` platform notes (incl. Linux
      build-from-source path)
- [x] `.gitignore` updated — `/bucky` + `/lib/`
- [ ] GitHub Actions: `go vet`, `staticcheck`, `gofmt -s -d`, `go build`
      on linux + darwin + windows runners — *not yet wired*

Verification (run on darwin/arm64, 2026-05-12):

- `go build -o bucky .` ✅
- `go vet ./...` ✅ clean
- `gofmt -s -l .` ✅ clean
- `go test -count=1 ./...` ✅ all pass
- `./bucky install -lib ./lib` ✅ downloads `whisper-v1.8.4-xcframework.zip`,
  extracts the macos-arm64_x86_64 universal Mach-O dylib to
  `./lib/libwhisper.dylib` (5.0 MiB, contains arm64 + x86_64 slices)
- `./bucky install -lib ./lib` second run ✅ idempotent
  ("whisper.cpp already installed at ./lib")
- `./bucky install -lib ./lib --os linux` ✅ fails with the documented
  "build from source per INSTALL.md" error
- `BUCKY_LIB=$(pwd)/lib ./bucky system` ✅ prints host + lib info

Out of scope: FFI calls, examples, audio decoding.

---

### PR #2 — FFI binding for the core whisper API  ✅ landed locally

Goal: `examples/hello` loads a model and transcribes a wav file end-to-end.

Tasks:

- [x] `pkg/whisper/whisper.go` — types/constants from `whisper.h`
      (whisper_token, whisper_pos, sampling strategy enum,
      alignment-heads preset, gretype, ahead/aheads, vad params,
      grammar element, opaque Context/State handles)
- [x] `pkg/whisper/model.go` — `whisper_init_from_file_with_params`,
      and the `whisper_model_n_*` accessor family + `model_type_readable`
- [x] `pkg/whisper/context.go` — `whisper_context_default_params`,
      `whisper_free`, and the `n_len/n_vocab/n_text_ctx/n_audio_ctx/
      is_multilingual` accessors
- [x] `pkg/whisper/params.go` — full `WhisperFullParams` struct mirror
      (304 bytes on darwin/arm64) plus `whisper_full_default_params`,
      `whisper_full_default_params_by_ref`, `whisper_free_params`
- [x] `pkg/whisper/full.go` — `whisper_full` blocking call,
      accepts pre-decoded f32 PCM samples
- [x] `pkg/whisper/segments.go` — `whisper_full_n_segments`,
      `whisper_full_get_segment_text`, `_t0`, `_t1`
- [x] `pkg/whisper/lang.go` — `whisper_lang_max_id`, `whisper_lang_id`,
      `whisper_lang_str`
- [x] `pkg/whisper/system.go` — `whisper_version`,
      `whisper_print_system_info`
- [x] `pkg/whisper/loader.go` — `Load(path)` that loads `libwhisper`
      (xcframework bundles ggml inside it on darwin) and runs each
      `loadXxxFuncs(lib)`
- [x] `pkg/utils/` — pure-Go `BytePtrFromString` / `BytePtrToString`
      wrappers over `golang.org/x/sys` (matches yzma)
- [x] `examples/hello/main.go` — loads tiny model, transcribes a bundled
      wav, prints text (inline minimal WAV reader; pkg/audio is PR #3)
- [x] `pkg/whisper/testmain_test.go` + `helpers_test.go` — env-gated
      setup (`BUCKY_LIB`, `BUCKY_TEST_MODEL`, `BUCKY_TEST_AUDIO`)
- [x] `pkg/whisper/params_test.go` — sizeof + by-ref/by-value field
      round-trip test for `WhisperFullParams`
- [x] `pkg/whisper/full_test.go` — end-to-end transcription test that
      asserts the JFK sample contains recognizable text
- [x] Update `Makefile` `test` target with `BUCKY_TEST_AUDIO`; add
      `make hello` convenience target
- [x] Bundle `samples/jfk.wav` (16 kHz mono 16-bit PCM, 344 KiB,
      from upstream whisper.cpp v1.8.4)
- [x] Wire `cmd/system.go` to call `whisper.Load` + `Version()` +
      `PrintSystemInfo()`

Verification (run on darwin/arm64, 2026-05-12):

- `go build ./...` ✅ clean
- `go vet ./...` ✅ clean
- `gofmt -s -l .` ✅ clean
- `go test -count=1 ./...` ✅ all pass (whisper FFI tests skip without
  env vars)
- `make download-whisper.cpp` + `make download-models` + `make test`
  ✅ all green on darwin/arm64 (TestWhisperFullParamsSize,
  TestContextDefaultParams, TestFullTranscribe pass)
- `go run ./examples/hello samples/jfk.wav` ✅ prints
  "And so my fellow Americans ask not what your country can do for you
  ask what you can do for your country."
- `BUCKY_LIB=$(pwd)/lib ./bucky system` ✅ prints whisper version
  (1.8.4) and the upstream `whisper_print_system_info` line

---

### PR #3 — Audio decoding + remaining transcription endpoints

Goal: accept real-world audio formats and surface the rest of the
useful whisper API (beam search, language detect, translate, segment
timing, per-token data).

Tasks:

- [ ] `pkg/audio/wav.go` — pure-Go WAV → f32 PCM (use
      `github.com/go-audio/wav` or hand-rolled)
- [ ] `pkg/audio/mp3.go` — `github.com/hajimehoshi/go-mp3`
- [ ] `pkg/audio/flac.go` — `github.com/mewkiz/flac`
- [ ] `pkg/audio/resample.go` — sinc / linear resampler to 16 kHz mono
- [ ] `pkg/audio/audio.go` — `Decode(io.Reader) ([]float32, error)`
      sniffs format and dispatches
- [ ] `pkg/whisper/params.go` — beam-search params, `n_threads`,
      `no_context`, `single_segment`, `print_*`
- [ ] `pkg/whisper/lang.go` — `whisper_lang_auto_detect`,
      `whisper_lang_id`, `whisper_lang_str`
- [ ] `pkg/whisper/full.go` — translate flag, `initial_prompt`
- [ ] `pkg/whisper/tokens.go` — per-segment token iteration,
      probabilities, timestamps
- [ ] Examples: `transcribe`, `translate`, `segments`
- [ ] CLI: `bucky whisper transcribe <file> [--lang xx] [--translate]`
- [ ] `MODELS.md` — list of recommended whisper GGUF/GGML model URLs
      (tiny, base, small, medium, large-v3, large-v3-turbo)
- [ ] `BENCHMARKS.md` skeleton

Verification:

- `examples/transcribe samples/jfk.mp3 --lang en` works
- `examples/translate samples/spanish.flac` returns English text
- `examples/segments samples/jfk.wav` prints timestamped segments
- All formats round-trip-able through `make test`

Out of scope: m4a / ogg / webm (documented as v2), streaming,
word-level timestamps, VAD, parallel decoders.

---

### PR #4 (optional, before Kronk integration) — VAD + parallel + word timing

Pull in once upstream features feel stable:

- [ ] `pkg/whisper/vad.go` — built-in VAD (whisper.cpp v1.7+)
- [ ] `pkg/whisper/state.go` — `whisper_full_parallel`, multiple states
- [ ] Word-level timestamps (`token_timestamps = true`,
      `dtw_token_timestamps`)
- [ ] `examples/streaming` — chunked / sliding-window pseudo-streaming

---

### PR #5 — Kronk integration (separate repo: kronk)

Once bucky's API is settled (post PR #3 minimum):

- [ ] Add `bucky` to kronk `go.mod`
- [ ] Extend `sdk/tools/libs` to also fetch whisper.cpp (or wire to
      `bucky install` programmatically)
- [ ] New `cmd/server/app/domain/transcribeapp/` mirroring `chatapp/`
      and `embedapp/`
- [ ] New SDK type: separate `WhisperPool` (do NOT merge into `Kronk`
      facade — Whisper is a different runtime)
- [ ] Wire JWT auth, rate limiter, Prometheus metrics, traces
- [ ] Multipart upload handler with configurable max size
      (default 25 MiB)
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
- Word-level timestamps via DTW — heavier, deferred.
- Real-time / streaming endpoint for kronk — separate kronk issue;
  do not fold into #565.
- m4a/ogg/webm — wait for user demand. Likely path is shipping
  `dr_libs` (header-only C) via purego rather than CGo or ffmpeg.

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
