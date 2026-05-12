package main

import (
	"fmt"
	"os"

	"github.com/ardanlabs/bucky/cmd"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:     "bucky",
		Usage:    "Bucky command line tool",
		Commands: buildCommands(),
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func buildCommands() []*cli.Command {
	return []*cli.Command{
		cmd.InstallCmd,
		cmd.SystemCmd,
		versionCmd,
		infoCmd,
	}
}

var versionCmd = &cli.Command{
	Name:  "version",
	Usage: "Show bucky version",
	Action: func(c *cli.Context) error {
		return runShowVersion(c)
	},
}

func runShowVersion(c *cli.Context) error {
	return showBuckyVersion()
}

func showBuckyVersion() error {
	fmt.Printf("bucky version %s\n", Version())
	return nil
}

func runShowInfo(c *cli.Context) error {
	cmd.ShowInfo(c)
	return showBuckyVersion()
}

var infoCmd = &cli.Command{
	Name:  "info",
	Usage: "Show bucky version",
	Action: func(c *cli.Context) error {
		return runShowInfo(c)
	},
}
