# bucky

Native Go binding for [whisper.cpp](https://github.com/ggml-org/whisper.cpp),
built without CGo using [purego](https://github.com/ebitengine/purego) and
[JupiterRider/ffi](https://github.com/JupiterRider/ffi). bucky is the
speech-to-text sibling of [hybridgroup/yzma](https://github.com/hybridgroup/yzma).

> Bucky is the squirrel from *The Emperor's New Groove* — he speaks in
> squeaks/whispers. Naming a Whisper binding after him is just good taste.

## Status

PR #1 (foundation) and PR #2 (core FFI binding) have landed. The CLI
(`bucky install`, `bucky system`, `bucky info`, `bucky version`), the
`pkg/loader` wrapper, and the `pkg/download` matrix for upstream whisper.cpp
release archives are in place. The `pkg/whisper` package now binds the core
whisper.cpp API — model loading, `whisper_full` transcription, segment
iteration, language helpers, and system info — using purego + jupiterrider/ffi
with no CGo. `examples/hello` transcribes a bundled `samples/jfk.wav`
end-to-end. Audio decoding for arbitrary formats lands in PR #3; see
[`PLAN.md`](./PLAN.md) for the full roadmap.

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

See [`INSTALL.md`](./INSTALL.md) for per-OS notes (Linux currently requires
building whisper.cpp from source).

## License

Apache-2.0 — see [`LICENSE`](./LICENSE).
