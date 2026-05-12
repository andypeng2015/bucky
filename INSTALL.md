# Installing whisper.cpp libraries for bucky

bucky loads `whisper.cpp` at runtime via [purego](https://github.com/ebitengine/purego)
and [jupiterrider/ffi](https://github.com/JupiterRider/ffi) — there is **no
CGo** in this repository. That means you need a prebuilt shared library on
disk before any FFI call will succeed.

Set `BUCKY_LIB` (or pass `-lib <path>`) to the directory that contains the
shared library. The expected filenames are:

| OS | Filename |
|---|---|
| linux / freebsd | `libwhisper.so` |
| darwin | `libwhisper.dylib` |
| windows | `whisper.dll` (plus `ggml*.dll` siblings) |

## macOS (arm64 / amd64)

```
make build
./bucky install -lib ./lib
```

This downloads the official `whisper-vX.Y.Z-xcframework.zip` from the
[whisper.cpp GitHub releases](https://github.com/ggml-org/whisper.cpp/releases)
and extracts the `macos-arm64_x86_64` slice of the universal Mach-O dylib
into `./lib/libwhisper.dylib`. Metal acceleration is included.

## Windows (amd64)

```
make build
.\bucky.exe install -lib .\lib
```

This downloads `whisper-bin-x64.zip` and extracts the `Release/*.dll`
files (`whisper.dll`, `ggml.dll`, `ggml-base.dll`, `ggml-cpu.dll`).

For CUDA builds use `-p cuda`; that downloads
`whisper-cublas-12.4.0-bin-x64.zip` instead. You must already have the
matching CUDA 12.4 runtime installed on the host.

## Linux (amd64 / arm64)

Upstream whisper.cpp does not publish prebuilt Linux binaries (as of
v1.8.4). `bucky install` will refuse with a clear error directing you here.
For now you must build whisper.cpp from source:

```
git clone https://github.com/ggml-org/whisper.cpp.git
cd whisper.cpp
git checkout v1.8.4
cmake -B build -DBUILD_SHARED_LIBS=ON
cmake --build build --config Release -j$(nproc)
mkdir -p ../bucky/lib
cp build/src/libwhisper.so ../bucky/lib/
cp build/ggml/src/libggml*.so ../bucky/lib/
```

Future versions of bucky may publish our own Linux bundles via a
`bucky-builder` companion repo (mirroring how
[`hybridgroup/llama-cpp-builder`](https://github.com/hybridgroup/llama-cpp-builder)
backs yzma); see `PLAN.md` parking lot.
