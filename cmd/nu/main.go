package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/dm-vev/nu/internal/app"
	"github.com/dm-vev/nu/internal/app/cli"
)

var (
	version   = "dev"
	commit    = ""
	buildDate = ""
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "get cwd: %v\n", err)
		os.Exit(1)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "get home: %v\n", err)
		os.Exit(1)
	}

	code := app.Run(ctx, app.Options{
		Args:   os.Args[1:],
		Env:    os.Environ(),
		CWD:    cwd,
		Home:   home,
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Version: cli.VersionInfo{
			Name:      "nu",
			Version:   version,
			Commit:    commit,
			BuildDate: buildDate,
		},
	})
	os.Exit(code)
}
