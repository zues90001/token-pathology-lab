// Pathology Lab — daily atlas + briefs publisher.
//
// Joins diagnostic findings with lab findings, runs the root-cause
// synthesizer + atlas publisher, and writes daily artefacts to disk.
package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/example/token-pathology-lab/internal/mimo"
	"github.com/example/token-pathology-lab/internal/storage"
	"github.com/example/token-pathology-lab/internal/treatment"
)

func main() {
	output := flag.String("output", "./atlas/", "directory for daily artefacts")
	diagDB := flag.String("diag-db", "./data/diagnostics.sqlite", "diagnostic db path")
	labDB := flag.String("lab-db", "./data/lab.sqlite", "lab db path")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	ctx := context.Background()

	diag, err := storage.OpenDiagnostics(*diagDB)
	if err != nil {
		logger.Error("diag open", "err", err)
		os.Exit(1)
	}
	defer diag.Close()

	lab, err := storage.OpenLab(*labDB)
	if err != nil {
		logger.Error("lab open", "err", err)
		os.Exit(1)
	}
	defer lab.Close()

	client, err := mimo.New(lab)
	if err != nil {
		logger.Error("mimo init", "err", err)
		os.Exit(1)
	}

	day := time.Now().UTC().Format("2006-01-02")
	dst := filepath.Join(*output, day)
	if err := os.MkdirAll(dst, 0o755); err != nil {
		logger.Error("mkdir", "err", err)
		os.Exit(1)
	}

	publisher := &treatment.Publisher{
		Client:    client,
		Diag:      diag,
		Lab:       lab,
		OutputDir: dst,
		Logger:    logger,
	}
	if err := publisher.Run(ctx); err != nil {
		logger.Error("publisher failed", "err", err)
		os.Exit(1)
	}
	logger.Info("published artefacts", "dir", dst)
}
