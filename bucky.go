// bucky lets you write Go applications that directly integrate whisper.cpp (https://github.com/ggml-org/whisper.cpp)
// for fully local automatic speech recognition (ASR) using hardware acceleration.
//
//   - Run any Whisper model on Linux, macOS, or Windows.
//   - Use any available hardware acceleration such as CUDA (https://en.wikipedia.org/wiki/CUDA),
//     Metal (https://en.wikipedia.org/wiki/Metal_(API)), or Vulkan (https://en.wikipedia.org/wiki/Vulkan)
//     for maximum performance.
//   - bucky uses the purego (https://github.com/ebitengine/purego) and ffi (https://github.com/JupiterRider/ffi)
//     packages so CGo is not needed.
//   - Works with the newest whisper.cpp releases so you can use the latest features and model support.
//
// bucky is the speech-to-text sibling of yzma (https://github.com/hybridgroup/yzma), which provides
// the same kind of FFI bindings for llama.cpp.
package main
