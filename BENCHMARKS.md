# Benchmarks

Performance numbers for `pkg/whisper`. Recorded on Apple M5 Max
(darwin/arm64) with the Metal backend baked into the upstream
`whisper-v1.8.4-xcframework.zip`. The Go benchmark and the upstream
ggml/memcpy helpers all run against the same `lib/libwhisper.dylib`.

Reproduce with:

```
make download-whisper.cpp           # populates ./lib
make download-models                # populates ~/models
make bench                          # BUCKY_BENCH_MODEL=ggml-tiny by default
```

## Methodology

- **Sample**: `samples/jfk.wav` — 11.0 s, 16 kHz mono 16-bit PCM (vendored
  from upstream whisper.cpp v1.8.4)
- **Driver**: `BenchmarkFullJFK` in `pkg/whisper/benchmark_test.go`. Greedy
  sampling, single-segment, no timestamp printing. One untimed warm-up
  iteration before `b.ResetTimer()` so Metal JIT/library init does not
  pollute the measurement.
- **Reported metrics**: `ns/op` (wall time per Full call), `audio_s`
  (length of the sample in seconds), and `rtf` (real-time factor =
  wall_seconds / audio_seconds; lower is faster, < 1 is faster than
  real-time playback).

## End-to-end transcription (greedy)

| Model | Backend | b.N | ns/op | audio_s | RTF |
|---|---|---:|---:|---:|---:|
| ggml-tiny | Metal | 10 | 27,960,167 | 11.00 | 0.0025 |

Run command:

```
BUCKY_LIB=$PWD/lib \
BUCKY_BENCH_MODEL=$HOME/models/ggml-tiny.bin \
BUCKY_TEST_AUDIO=$PWD/samples/jfk.wav \
go test -bench=BenchmarkFullJFK -benchtime=10x -run='^$' ./pkg/whisper/
```

The first un-timed warm-up dominates total wall time (~5–6 s) because of
Metal library compilation; warm runs are ~28 ms for 11 s of audio. To
record numbers across `tiny`, `base`, `small`, etc., re-run with a
different `BUCKY_BENCH_MODEL`.

## Built-in upstream micro-benchmarks

`whisper.cpp` exposes two helpers that we surface as
`whisper.BenchMemcpyStr` and `whisper.BenchGGMLMulMatStr`. These are
useful for comparing backends or hosts without loading a model.

### `whisper_bench_memcpy_str(4)` — Apple M5 Max

```
memcpy:   61.08 GB/s (heat-up)
memcpy:   69.29 GB/s ( 1 thread)
memcpy:   67.17 GB/s ( 1 thread)
memcpy:  119.11 GB/s ( 2 thread)
memcpy:  159.51 GB/s ( 3 thread)
memcpy:  167.28 GB/s ( 4 thread)
```

### `whisper_bench_ggml_mul_mat_str(4)` — selected sizes

| Size | Q4_0 | Q8_0 | F16 | F32 |
|---|---:|---:|---:|---:|
| 256x256   | 112.7 GFLOPS | 229.1 GFLOPS | 201.1 GFLOPS | 138.0 GFLOPS |
| 512x512   | 133.8 GFLOPS | 370.7 GFLOPS | 306.7 GFLOPS | 175.6 GFLOPS |
| 1024x1024 | 138.1 GFLOPS | 428.2 GFLOPS | 359.3 GFLOPS | 183.0 GFLOPS |
| 2048x2048 | 139.2 GFLOPS | 415.3 GFLOPS | 355.2 GFLOPS | 165.8 GFLOPS |
| 4096x4096 | 139.4 GFLOPS | 383.2 GFLOPS | 324.0 GFLOPS | 153.9 GFLOPS |

Full output is produced by the `BenchMemcpyStr` / `BenchGGMLMulMatStr`
wrappers — invoke them directly from a small driver program if you want
the complete table.
