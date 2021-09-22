package app

import (
	"context"
	"fmt"
	"io"

	"github.com/DaniilGo/challenge/internal/merge"
	"github.com/DaniilGo/challenge/internal/source"
)

type App struct {
	source source.Source
	merger merge.Merger
	output io.Writer
}

func NewApp(source source.Source, merger merge.Merger, output io.Writer) *App {
	return &App{
		source: source,
		merger: merger,
		output: output,
	}
}

func (a *App) Run(ctx context.Context) error {
	r, err := a.source.GetReadCloser(ctx)
	if err != nil {
		return fmt.Errorf("failed to get read closer: %w", err)
	}

	defer func() {
		_ = r.Close()
	}()

	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("failed to read all: %w", err)
	}

	merged, err := a.merger.GetMerged(data)
	if err != nil {
		return fmt.Errorf("failed to get merged: %w", err)
	}

	if _, err = a.output.Write(merged); err != nil {
		return fmt.Errorf("failed to write: %w", err)
	}

	return nil
}
