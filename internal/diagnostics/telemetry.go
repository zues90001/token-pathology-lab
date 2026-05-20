// Package diagnostics handles opt-in telemetry ingest + complaint scrape.
//
// The handlers in this package are deterministic. The MiMo-driven
// pattern-classifier lives in the same package but is invoked by the
// publisher binary, not by the http listener.
package diagnostics

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/example/token-pathology-lab/internal/storage"
)

// telemetryPayload is the shape opt-in tenants POST.
type telemetryPayload struct {
	TenantID         string `json:"tenant_id"`
	PromptHash       string `json:"prompt_hash"`
	Model            string `json:"model"`
	PromptTokens     int    `json:"prompt_tokens"`
	CompletionTokens int    `json:"completion_tokens"`
	ReasoningTokens  int    `json:"reasoning_tokens"`
	TotalTokens      int    `json:"total_tokens"`
	RetryCount       int    `json:"retry_count"`
	TaskClass        string `json:"task_class"`
}

// NewTelemetryHandler returns the HTTP handler for /telemetry.
func NewTelemetryHandler(store *storage.Diagnostics, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		if r.Header.Get("X-Optin") != "1" {
			http.Error(w, "missing X-Optin: 1", http.StatusBadRequest)
			return
		}
		var p telemetryPayload
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// Strip identifiers at ingest. Anything we store is hashed.
		tenantHash := sha256Hex(p.TenantID)
		err := store.AppendTelemetry(r.Context(),
			tenantHash, p.PromptHash, p.Model, p.TaskClass,
			p.PromptTokens, p.CompletionTokens, p.ReasoningTokens,
			p.TotalTokens, p.RetryCount,
		)
		if err != nil {
			logger.Warn("append telemetry failed", "err", err)
			http.Error(w, "store failed", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	})
}

func sha256Hex(in string) string {
	sum := sha256.Sum256([]byte(in))
	return hex.EncodeToString(sum[:])
}
