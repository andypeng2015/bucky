# bucky

Native Go binding for [whisper.cpp](https://github.com/ggml-org/whisper.cpp),
built without CGo using [purego](https://github.com/ebitengine/purego) and
[JupiterRider/ffi](https://github.com/JupiterRider/ffi). bucky is the
speech-to-text sibling of [hybridgroup/yzma](https://github.com/hybridgroup/yzma).

> Bucky is the squirrel from *The Emperor's New Groove* — he speaks in
> squeaks/whispers. Naming a Whisper binding after him is just good taste.

## Status

PR #1 (foundation), PR #2 (core FFI binding), and PR #3 (audio decoding +
remaining transcription endpoints) have landed. Bucky now provides:

- A CLI: `bucky install`, `bucky system`, `bucky info`, `bucky version`,
  `bucky whisper transcribe …`
- `pkg/loader` + `pkg/download` for fetching the upstream whisper.cpp
  release archives
- `pkg/whisper` — pure-Go FFI bindings (no CGo) to the core whisper.cpp
  API: model loading, `whisper_full` / `whisper_full_parallel`, segment
  + per-token iteration, language detection, translation, beam search
- `pkg/audio` — pure-Go decoders for WAV (8/16/24/32-bit PCM and
  32-bit float), MP3 (via go-mp3), FLAC (via mewkiz/flac), with
  channel downmix and linear resampling to 16 kHz
- Examples: `hello`, `transcribe`, `translate`, `segments`

VAD, parallel decoders, word-level timestamps via DTW, and streaming are
deferred to PR #4. See [`PLAN.md`](./PLAN.md) for the full roadmap.

## Quickstart

```
make build
./bucky install -lib ./lib
export BUCKY_LIB=$(pwd)/lib
./bucky system

# transcribe the bundled JFK sample with a tiny model
make download-models
BUCKY_TEST_MODEL=$HOME/models/ggml-tiny.bin \
    go run ./examples/hello samples/jfk.wav
```

Or via the CLI:

```
./bucky whisper transcribe -m $HOME/models/ggml-tiny.bin --segments samples/jfk.wav
```

See [`INSTALL.md`](./INSTALL.md) for per-OS notes (Linux currently requires
building whisper.cpp from source) and [`MODELS.md`](./MODELS.md) for
recommended whisper models.

## License

Apache-2.0 — see [`LICENSE`](./LICENSE).
