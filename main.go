package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/Softorize/yoy/cmd"
	"github.com/Softorize/yoy/internal/config"
	yoyerrors "github.com/Softorize/yoy/internal/errors"
)

func main() {
	var cli cmd.CLI
	kongCtx := kong.Parse(&cli,
		kong.Name("yoy"),
		kong.Description("Yahoo Mail CLI"),
		kong.UsageOnError(),
	)

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not load config: %v\n", err)
		cfg = &config.Config{}
	}

	ctx := cmd.NewContext(&cli, cfg)
	defer ctx.Close()

	err = kongCtx.Run(ctx)
	if err != nil {
		hint := yoyerrors.HintFrom(err)
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		if hint != "" {
			fmt.Fprintf(os.Stderr, "Hint: %s\n", hint)
		}
		os.Exit(yoyerrors.ExitCodeFrom(err))
	}
}
