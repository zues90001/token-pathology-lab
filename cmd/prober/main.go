// Pathology Lab — depth-graded prober.
//
// Walks the synthetic corpus, runs each task at 5 depth levels via MiMo,
// captures the outputs, and triggers the quality judge.
package main

import (
	"context"
	"flag"
	"log/slog"
	"os"

	"github.com/example/token-pathology-lab/internal/lab"
	"github.com/example/token-pathology-lab/internal/mimo"
	"github.com/example/token-pathology-lab/internal/storage"
)

func main() {
	domain := flag.String("domain", "code", "synthetic corpus domain")
	tasks := flag.Int("tasks", 200, "tasks per probe run")
	dbPath := flag.String("db", "./data/lab.sqlite", "lab db path")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	ctx := context.Background()

	store, err := storage.OpenLab(*dbPath)
	if err != nil {
		logger.Error("storage open", "err", err)
		os.Exit(1)
	}
	defer store.Close()

	client, err := mimo.New(store)
	if err != nil {
		logger.Error("mimo client init", "err", err)
		os.Exit(1)
	}

	runner := &lab.Runner{
		Client: client,
		Store:  store,
		Logger: logger,
	}
	if err := runner.Run(ctx, *domain, *tasks); err != nil {
		logger.Error("prober run failed", "err", err)
		os.Exit(1)
	}
}
