package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ardanlabs/bucky/pkg/download"
	"github.com/ardanlabs/bucky/pkg/whisper"
	"github.com/urfave/cli/v2"
)

// ModelCmd manages whisper model files (download / inspect).
var ModelCmd = &cli.Command{
	Name:  "model",
	Usage: "Manage whisper models",
	Subcommands: []*cli.Command{
		modelGetCmd,
		modelInfoCmd,
		modelListCmd,
	},
}

var modelGetCmd = &cli.Command{
	Name:      "get",
	Usage:     "Download a whisper model from a URL or short name",
	ArgsUsage: "[short-name]",
	Description: `Download a whisper model into the local models directory.

The argument may be either a short name from the bundled catalog (see
"bucky model list") or omitted in favor of -u/--url to fetch from any
URL accepted by hashicorp/go-getter (https://, file://, s3:// etc.).

Examples:
  bucky model get tiny
  bucky model get -o /tmp/models base.en
  bucky model get -u https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-small.bin`,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "url",
			Aliases: []string{"u"},
			Usage:   "explicit URL to download (overrides any short-name argument)",
		},
		&cli.StringFlag{
			Name:        "output",
			Aliases:     []string{"o"},
			Usage:       "directory to save the model file into",
			Value:       download.DefaultModelsDir(),
			DefaultText: "~/models",
		},
		&cli.BoolFlag{
			Name:    "yes",
			Aliases: []string{"y"},
			Usage:   "create the output directory without prompting",
			Value:   false,
		},
		&cli.BoolFlag{
			Name:  "show-progress",
			Usage: "print download progress to stdout",
			Value: true,
		},
	},
	Action: func(c *cli.Context) error {
		return runModelGet(c)
	},
}

func runModelGet(c *cli.Context) error {
	url := c.String("url")
	output := c.String("output")
	autoYes := c.Bool("yes")
	showProgress := c.Bool("show-progress")

	if url == "" {
		if c.NArg() < 1 {
			return fmt.Errorf("provide either -u <url> or a short model name (see `bucky model list`)")
		}
		name := c.Args().First()
		entry, ok := whisperModelByName(name)
		if !ok {
			return fmt.Errorf("unknown model %q (run `bucky model list` to see known names)", name)
		}
		url = entry.URL
	}

	if _, err := os.Stat(output); os.IsNotExist(err) {
		if !autoYes {
			fmt.Printf("Directory %s does not exist.\n", output)
			fmt.Print("Would you like to create it? [y/N]: ")
			var response string
			fmt.Scanln(&response)
			response = strings.ToLower(strings.TrimSpace(response))
			if response != "y" && response != "yes" {
				fmt.Println("Download cancelled.")
				return nil
			}
		}
		if err := os.MkdirAll(output, 0o755); err != nil {
			return fmt.Errorf("create output directory: %w", err)
		}
		fmt.Printf("Created directory %s\n", output)
	}

	fmt.Printf("Downloading %s into %s ...\n", url, output)

	if !showProgress {
		download.ProgressTracker = nil
	}

	if err := download.GetModel(url, output); err != nil {
		return fmt.Errorf("download model: %w", err)
	}

	fmt.Println("Download completed successfully.")
	return nil
}

var modelInfoCmd = &cli.Command{
	Name:      "info",
	Usage:     "Show information about a downloaded whisper model",
	ArgsUsage: "<model-path>",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "model",
			Aliases: []string{"m"},
			Usage:   "path to the GGML whisper model file",
		},
		&cli.StringFlag{
			Name:    "lib",
			Aliases: []string{"l"},
			Usage:   "path to the directory containing libwhisper",
			EnvVars: []string{"BUCKY_LIB"},
		},
	},
	Action: func(c *cli.Context) error {
		return runModelInfo(c)
	},
}

func runModelInfo(c *cli.Context) error {
	model := c.String("model")
	if model == "" && c.NArg() > 0 {
		model = c.Args().First()
	}
	if model == "" {
		return fmt.Errorf("provide a model path via -m or as an argument")
	}
	if _, err := os.Stat(model); err != nil {
		return fmt.Errorf("model file: %w", err)
	}

	libPath := c.String("lib")
	if libPath == "" {
		return fmt.Errorf("missing -lib flag or BUCKY_LIB env var")
	}
	if err := whisper.Load(libPath); err != nil {
		return fmt.Errorf("load whisper library: %w", err)
	}

	if err := whisper.Init(libPath); err != nil {
		return fmt.Errorf("init whisper library: %w", err)
	}

	cparams := whisper.ContextDefaultParams()
	ctx, err := whisper.InitFromFileWithParams(model, cparams)
	if err != nil {
		return fmt.Errorf("init model: %w", err)
	}
	defer whisper.Free(ctx)

	abs, _ := filepath.Abs(model)
	fmt.Printf("file:               %s\n", abs)
	fmt.Printf("type:               %s (id %d)\n", whisper.ModelTypeReadable(ctx), whisper.ModelType(ctx))
	fmt.Printf("vocab:              %d\n", whisper.ModelNVocab(ctx))
	fmt.Printf("ftype:              %d\n", whisper.ModelFtype(ctx))
	fmt.Printf("audio_ctx:          %d\n", whisper.ModelNAudioCtx(ctx))
	fmt.Printf("audio_state:        %d\n", whisper.ModelNAudioState(ctx))
	fmt.Printf("audio_head:         %d\n", whisper.ModelNAudioHead(ctx))
	fmt.Printf("audio_layer:        %d\n", whisper.ModelNAudioLayer(ctx))
	fmt.Printf("text_ctx:           %d\n", whisper.ModelNTextCtx(ctx))
	fmt.Printf("text_state:         %d\n", whisper.ModelNTextState(ctx))
	fmt.Printf("text_head:          %d\n", whisper.ModelNTextHead(ctx))
	fmt.Printf("text_layer:         %d\n", whisper.ModelNTextLayer(ctx))
	fmt.Printf("mels:               %d\n", whisper.ModelNMels(ctx))
	fmt.Printf("multilingual:       %t\n", whisper.IsMultilingual(ctx))
	return nil
}

var modelListCmd = &cli.Command{
	Name:  "list",
	Usage: "List the bundled catalog of well-known whisper models",
	Action: func(c *cli.Context) error {
		return runModelList(c)
	},
}

func runModelList(c *cli.Context) error {
	names := make([]string, 0, len(whisperCatalog))
	for n := range whisperCatalog {
		names = append(names, n)
	}
	sort.Strings(names)

	fmt.Printf("%-22s %-10s %s\n", "NAME", "SIZE", "URL")
	for _, n := range names {
		e := whisperCatalog[n]
		fmt.Printf("%-22s %-10s %s\n", n, e.Size, e.URL)
	}
	fmt.Println()
	fmt.Println("Use `bucky model get <name>` to download into ~/models (override with -o).")
	return nil
}

// catalogEntry is one row in the bundled whisper-model catalog.
type catalogEntry struct {
	URL  string
	Size string
}

// whisperModelByName returns the catalog entry for a short name, also
// accepting the full ggml-<name>.bin filename for convenience.
func whisperModelByName(name string) (catalogEntry, bool) {
	if e, ok := whisperCatalog[name]; ok {
		return e, true
	}
	// allow "ggml-tiny.bin" -> "tiny"
	trimmed := strings.TrimSuffix(strings.TrimPrefix(name, "ggml-"), ".bin")
	if e, ok := whisperCatalog[trimmed]; ok {
		return e, true
	}
	return catalogEntry{}, false
}

// whisperCatalog is the curated set of well-known whisper models. URLs are
// pinned to the upstream HuggingFace mirror. Sizes are nominal (compressed
// download sizes vary slightly).
var whisperCatalog = map[string]catalogEntry{
	"tiny":           {URL: "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-tiny.bin", Size: "75 MB"},
	"tiny.en":        {URL: "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-tiny.en.bin", Size: "75 MB"},
	"base":           {URL: "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-base.bin", Size: "142 MB"},
	"base.en":        {URL: "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-base.en.bin", Size: "142 MB"},
	"small":          {URL: "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-small.bin", Size: "466 MB"},
	"small.en":       {URL: "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-small.en.bin", Size: "466 MB"},
	"medium":         {URL: "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-medium.bin", Size: "1.5 GB"},
	"medium.en":      {URL: "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-medium.en.bin", Size: "1.5 GB"},
	"large-v3":       {URL: "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-large-v3.bin", Size: "2.9 GB"},
	"large-v3-turbo": {URL: "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-large-v3-turbo.bin", Size: "1.5 GB"},
	"silero-vad":     {URL: "https://huggingface.co/ggml-org/whisper-vad/resolve/main/ggml-silero-v5.1.2.bin", Size: "0.9 MB"},
}
