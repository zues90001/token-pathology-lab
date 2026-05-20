// Pathology Lab — opt-in telemetry collector.
//
// Listens for tenant-emitted telemetry over HTTPS. Strips identifiers and
// records only token counts + structural metadata. The collector is
// intentionally separate from the lab and treatment binaries: the
// telemetry stream and the synthetic-corpus stream never share storage.
package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"

	"github.com/example/token-pathology-lab/internal/diagnostics"
	"github.com/example/token-pathology-lab/internal/storage"
)

func main() {
	listen := flag.String("listen", ":8090", "address to bind")
	dbPath := flag.String("db", "./data/diagnostics.sqlite", "diagnostic db")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	store, err := storage.OpenDiagnostics(*dbPath)
	if err != nil {
		logger.Error("storage open", "err", err)
		os.Exit(1)
	}
	defer store.Close()

	mux := http.NewServeMux()
	mux.Handle("/telemetry", diagnostics.NewTelemetryHandler(store, logger))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	logger.Info("collector listening", "addr", *listen, "db", *dbPath)
	if err := http.ListenAndServe(*listen, mux); err != nil {
		logger.Error("listen failed", "err", err)
		os.Exit(1)
	}
}
