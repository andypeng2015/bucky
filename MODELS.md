# Whisper Models

Bucky loads any GGML-format whisper model. The official set lives at
<https://huggingface.co/ggerganov/whisper.cpp>. Download the `.bin` files
directly with `curl` (or use `make download-models`).

## Recommended set

| Model | Size | Speed (M-class) | Quality | When to pick |
|---|---|---|---|---|
| `ggml-tiny.bin` | 75 MB | very fast | low | unit tests, smoke tests, low-resource hardware |
| `ggml-tiny.en.bin` | 75 MB | very fast | low | English-only smoke tests; slightly better than tiny on English |
| `ggml-base.bin` | 142 MB | fast | acceptable | quick transcripts, casual notes |
| `ggml-base.en.bin` | 142 MB | fast | acceptable | English-only baseline |
| `ggml-small.bin` | 466 MB | medium | good | default for most production use |
| `ggml-medium.bin` | 1.5 GB | slow | very good | when accuracy matters and latency is tolerable |
| `ggml-large-v3.bin` | 2.9 GB | slowest | best | offline batch jobs; archival quality |
| `ggml-large-v3-turbo.bin` | 1.5 GB | fast | very good | best speed/quality trade-off; recommended default once stable |

`-q5_0`, `-q5_1`, `-q8_0` etc. variants are quantized and smaller; pick one
when memory is tight. They cost ~1–3% accuracy.

`-en` variants are English-only and reject `--translate`. Use a non-`-en`
("multilingual") model for translation or non-English input.

## Direct download URLs

Replace `<name>` with one of the model names above:

```
https://huggingface.co/ggerganov/whisper.cpp/resolve/main/<name>
```

## make download-models

```sh
MODELS_DIR=$HOME/models make download-models
```

Fetches `ggml-tiny.bin` and `ggml-base.en.bin` into `$MODELS_DIR`. Edit
`Makefile` to add more.

## Verifying a model works

```sh
BUCKY_LIB=$(pwd)/lib \
BUCKY_TEST_MODEL=$HOME/models/ggml-tiny.bin \
go run ./examples/hello samples/jfk.wav
```

Should print:

> And so my fellow Americans ask not what your country can do for you ask
> what you can do for your country.
