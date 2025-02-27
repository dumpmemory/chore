package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"text/scanner"
	"unicode"

	"github.com/9seconds/chore/internal/cli"
	"github.com/alecthomas/kong"
)

var version = "dev"

const fileDefaultPermission = 0o600

var CLI struct {
	Debug   bool             `short:"d" env:"CHORE_DEBUG" help:"Run in debug mode."`
	Version kong.VersionFlag `short:"V" help:"Show version."`

	Show           CliCmdShow        `cmd:"" aliases:"s" help:"Show details on namespaces or scripts."`
	EditScript     CliCmdEditScript  `cmd:"" aliases:"e" help:"Edit chore script."`
	EditConfig     CliCmdEditConfig  `cmd:"" aliases:"c" help:"Edit chore script config."`
	Run            CliCmdRun         `cmd:"" aliases:"r" help:"Run chore script."`
	FishCompletion CliFishCompletion `cmd:"" help:"Generate fish shell completion."`
}

type editorCommand struct {
	Editor cli.Editor `short:"e" help:"Editor to use."`

	Namespace cli.Namespace `arg:"" help:"Script namespace."`
	Script    string        `arg:"" help:"Script name."`
}

func (e *editorCommand) Open(ctx context.Context, path string, templateContent []byte) error {
	editor, err := e.Editor.Value()
	if err != nil {
		return fmt.Errorf("cannot initialize editor: %w", err)
	}

	_, err = os.Stat(path)

	switch {
	case errors.Is(err, fs.ErrNotExist):
		if err := os.WriteFile(path, templateContent, fileDefaultPermission); err != nil {
			return fmt.Errorf("cannot populate file with a template content: %w", err)
		}
	case err != nil:
		return fmt.Errorf("cannot stat file: %w", err)
	}

	cmd := exec.CommandContext(ctx, editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (e *editorCommand) RemoveIfEmpty(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, fmt.Errorf("cannot open file: %w", err)
	}

	defer file.Close()

	text := scanner.Scanner{}
	text.Init(file)

	for tok := text.Scan(); tok != scanner.EOF; tok = text.Scan() {
		if !unicode.IsSpace(tok) {
			return false, nil
		}
	}

	if err := os.Remove(path); err != nil {
		return false, fmt.Errorf("cannot remove %s: %w", path, err)
	}

	rootDir := filepath.Dir(path)

	if items, err := os.ReadDir(rootDir); err == nil && len(items) == 0 {
		return true, os.RemoveAll(rootDir)
	}

	return true, nil
}
