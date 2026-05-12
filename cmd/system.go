package cmd

import (
	"fmt"
	"runtime"

	"github.com/ardanlabs/bucky/pkg/whisper"
	"github.com/urfave/cli/v2"
)

// SystemCmd shows information about the host environment and the loaded
// whisper.cpp library.
var SystemCmd = &cli.Command{
	Name:  "system",
	Usage: "Show whisper.cpp / system information",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "lib",
			Aliases: []string{"l"},
			Usage:   "path to whisper.cpp compiled library files",
			EnvVars: []string{"BUCKY_LIB"},
		},
	},
	Action: func(c *cli.Context) error {
		return runSystemInfo(c)
	},
}

func runSystemInfo(c *cli.Context) error {
	libPath := c.String("lib")

	fmt.Println("-- Host --")
	fmt.Printf("os:   %s\n", runtime.GOOS)
	fmt.Printf("arch: %s\n", runtime.GOARCH)
	fmt.Printf("cpus: %d\n", runtime.NumCPU())
	fmt.Println()

	fmt.Println("-- Library --")
	if libPath == "" {
		fmt.Println("BUCKY_LIB not set; pass -lib or set the env var")
		return nil
	}
	fmt.Println("path:", libPath)

	if err := whisper.Load(libPath); err != nil {
		return fmt.Errorf("failed to load whisper.cpp from %s: %w", libPath, err)
	}

	fmt.Println("version:", whisper.Version())
	fmt.Println()
	fmt.Println("-- Whisper system info --")
	fmt.Println(whisper.PrintSystemInfo())
	return nil
}
