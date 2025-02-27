package main

import (
	"context"
	"errors"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/9seconds/chore/internal/cli"
	"github.com/9seconds/chore/internal/commands"
	"github.com/alecthomas/kong"
)

func main() {
	cliCtx := kong.Parse(
		&CLI,
		kong.Name("chore"),
		kong.Description("Execution environment for a small helper scripts."),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact:   true,
			Summary:   true,
			Tree:      true,
			FlagsLast: true,
		}),
		kong.Vars{
			"version": version,
		})

	log.SetFlags(log.Lshortfile | log.LstdFlags)

	if !CLI.Debug {
		log.SetOutput(io.Discard)
	}

	notifyCtx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go func() {
		<-notifyCtx.Done()
		log.Println("application context is closed")
	}()

	rootCtx := cli.NewContext(notifyCtx)
	err := cliCtx.Run(rootCtx)

	rootCtx.Close()
	cancel()

	var exitErr commands.ExitError

	if errors.As(err, &exitErr) {
		os.Exit(exitErr.Code())
	}

	cliCtx.FatalIfErrorf(err)
}
