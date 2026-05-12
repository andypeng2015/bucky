package cmd

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

// ShowInfo prints the bucky banner and a short tagline.
func ShowInfo(c *cli.Context) error {
	fmt.Println(logo)
	fmt.Println()
	fmt.Println("Local speech-to-text in Go using whisper.cpp with hardware acceleration")

	return nil
}

const logo = `
 ____  __ __  _____  __  _ __ __ 
|    \|  |  ||     ||  |/ ||  |  |
|  o  )  |  ||   __||  ' / |  |  |
|     )  |  ||  |__ |    \ |  ~  |
|  O  )  :  ||   __||     \|___, |
|     |     ||  |__ |  .  ||     |
|_____|\__,_||_____||__|\_||____/`
