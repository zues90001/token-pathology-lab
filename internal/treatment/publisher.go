// Package treatment joins diagnostic + lab findings, runs the root-cause
// synthesizer, and writes daily artefacts to disk.
package treatment

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/example/token-pathology-lab/internal/mimo"
	"github.com/example/token-pathology-lab/internal/storage"
)

const synthSystem = `You are a root-cause synthesizer. You receive:
- a batch of diagnostic findings for one task class (real-world signal)
- the lab finding for the same task class (controlled experiment)

You produce:
- a proposed REMEDIATION_PROMPT_PREFIX
- a proposed MODEL_TRAINING_HINT
- an estimated ECOSYSTEM_SAVINGS_TOK_PER_MONTH

Output strict JSON. If diagnostic frequency is unknown, set
ECOSYSTEM_SAVINGS_TOK_PER_MONTH = null.`

const publisherSystem = `You are the daily Atlas Publisher. You ship four
artefacts per run:
  1. mimo_team_brief_md  — root cause + recommended training adjustment
  2. tenant_tips         — per-tenant remediation tip (one per opt-in)
  3. community_digest_md — anonymised aggregate (no prompt text)
  4. atlas_row           — single CSV-row record for the open atlas

You receive: today's diagnostic + lab + synthesis findings.
Output strict JSON:
{
  "mimo_team_brief_md": "...",
  "tenant_tips": { "<tenant_id>": "...", ... },
  "community_digest_md": "...",
  "atlas_row": { ... }
}`

// Publisher is the daily orchestrator.
type Publisher struct {
	Client    *mimo.Client
	Diag      *storage.Diagnostics
	Lab       *storage.Lab
	OutputDir string
	Logger    *slog.Logger
}

// Run executes one daily publication cycle.
func (p *Publisher) Run(ctx context.Context) error {
	// In production this loads today's findings from p.Diag and p.Lab.
	// This skeleton emits a deterministic placeholder bundle so reviewers
	// can see the pipeline shape without a live MiMo key.
	bundle := map[string]any{
		"task_class": "json_classification_short",
		"diagnostic": map[string]any{
			"samples_real":          1450,
			"complaint_correlation": 0.81,
		},
		"lab": map[string]any{
			"min_viable_depth_pct":         25,
			"quality_curve":                map[string]float64{"10": 0.62, "25": 0.95, "50": 0.96, "75": 0.96, "100": 0.97},
			"failure_mode_below_threshold": "At 10% depth model omits 1 of 3 required JSON fields.",
		},
	}
	bundleJSON, _ := json.Marshal(bundle)

	synth, err := p.Client.Chat(ctx, mimo.ChatRequest{
		Agent:    "root_cause_synthesizer",
		System:   synthSystem,
		User:     string(bundleJSON),
		Temp:     0,
		MaxToks:  1500,
		JSONMode: true,
	})
	if err != nil {
		p.Logger.Warn("synth call failed (skeleton continues)", "err", err)
	}

	pubBundle := map[string]any{
		"task_class":      "json_classification_short",
		"diagnostic":      bundle["diagnostic"],
		"lab":             bundle["lab"],
		"synth_response":  synth,
	}
	pubJSON, _ := json.Marshal(pubBundle)

	out, err := p.Client.Chat(ctx, mimo.ChatRequest{
		Agent:    "atlas_publisher",
		System:   publisherSystem,
		User:     string(pubJSON),
		Temp:     0.2,
		MaxToks:  4000,
		JSONMode: true,
	})
	if err != nil {
		p.Logger.Warn("publisher call failed (skeleton continues)", "err", err)
	}

	if out != "" {
		if err := os.WriteFile(filepath.Join(p.OutputDir, "raw_publisher_output.json"), []byte(out), 0o644); err != nil {
			return fmt.Errorf("write raw publisher output: %w", err)
		}
	}
	return nil
}

// AtlasEntry is the published, training-ready schema entry.
type AtlasEntry struct {
	Title        string
	Category     string
	BeforePrompt string
	AfterPrompt  string
	TokenDelta   int
}

// ValidateEntry checks the minimum schema before publishing.
func ValidateEntry(e *AtlasEntry) error {
	if e.Title == "" {
		return errFieldRequired("title")
	}
	if e.Category == "" {
		return errFieldRequired("category")
	}
	return nil
}

func slugify(s string) string {
	out := []rune{}
	last := '-'
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			out = append(out, r)
			last = r
		} else if r >= 'A' && r <= 'Z' {
			out = append(out, r+32)
			last = r + 32
		} else if last != '-' {
			out = append(out, '-')
			last = '-'
		}
	}
	res := string(out)
	for len(res) > 0 && res[0] == '-' {
		res = res[1:]
	}
	for len(res) > 0 && res[len(res)-1] == '-' {
		res = res[:len(res)-1]
	}
	return res
}

type fieldError string

func (e fieldError) Error() string { return string(e) + " is required" }
func errFieldRequired(f string) error { return fieldError(f) }
