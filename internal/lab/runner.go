// Package lab implements the depth-graded prober + quality judge.
//
// The runner walks a synthetic corpus, fires MiMo for each (task, depth)
// pair, persists the output, and asks MiMo to judge plateau depth.
package lab

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/example/token-pathology-lab/internal/mimo"
	"github.com/example/token-pathology-lab/internal/storage"
)

const proberSystem = `You are a controlled depth-graded inference runner.
You will execute the USER prompt under a reasoning-depth budget hint.

You may use chain-of-thought up to the depth budget specified. Once the
budget is exhausted, you must terminate reasoning and emit the answer.

Depth budget mapping:
- 10%  → at most 1 short reasoning step
- 25%  → at most 3 reasoning steps
- 50%  → unrestricted but bias toward concision
- 75%  → standard reasoning, prefer step-by-step
- 100% → full reasoning chain, no concision pressure

Always terminate with the requested output schema. Never explain the
budget decision in the output.`

const judgeSystem = `You are a quality judge. You receive five outputs of
the same task, each produced under a different reasoning-depth budget
(10/25/50/75/100%).

Compare them on:
- correctness (factual or schema accuracy)
- completeness (did the answer cover all required fields?)
- clarity (would a reader of the output understand the result?)

Output strict JSON:
{
  "min_viable_depth_pct": 10|25|50|75|100,
  "quality_curve": { "10": <0-1>, "25": <0-1>, "50": <0-1>, "75": <0-1>, "100": <0-1> },
  "failure_mode_below_threshold": "<one sentence>",
  "rationale": "<two sentences>"
}

Be conservative. If two depths produce equally good output, choose the
lower depth.`

// Task is a single synthetic-corpus item.
type Task struct {
	ID        string
	Class     string
	Domain    string
	UserPrompt string
}

// Verdict is the judge's output.
type Verdict struct {
	MinViableDepthPct          int                `json:"min_viable_depth_pct"`
	QualityCurve               map[string]float64 `json:"quality_curve"`
	FailureModeBelowThreshold  string             `json:"failure_mode_below_threshold"`
	Rationale                  string             `json:"rationale"`
}

// Runner orchestrates the depth probe + judge for a domain slice.
type Runner struct {
	Client *mimo.Client
	Store  *storage.Lab
	Logger *slog.Logger
}

// Run probes a fixed number of tasks in the requested domain.
func (r *Runner) Run(ctx context.Context, domain string, n int) error {
	tasks := SyntheticCorpus(domain, n)
	for _, t := range tasks {
		outputs := map[int]string{}
		for _, depth := range []int{10, 25, 50, 75, 100} {
			out, err := r.Client.Chat(ctx, mimo.ChatRequest{
				Agent:    fmt.Sprintf("depth_prober_%d", depth),
				System:   proberSystem + fmt.Sprintf("\n\nReasoning depth budget: %d%%", depth),
				User:     t.UserPrompt,
				Temp:     0,
				MaxToks:  3000,
				JSONMode: false,
			})
			if err != nil {
				r.Logger.Warn("probe failed", "task", t.ID, "depth", depth, "err", err)
				continue
			}
			outputs[depth] = out
			_ = r.Store.AppendProbe(ctx, t.ID, t.Class, t.Domain, depth, 0, 0, out)
		}
		if len(outputs) < 5 {
			continue
		}
		judgeInput, _ := json.Marshal(outputs)
		raw, err := r.Client.Chat(ctx, mimo.ChatRequest{
			Agent:    "quality_judge",
			System:   judgeSystem,
			User:     fmt.Sprintf(`Task class: %s\nOutputs by depth: %s`, t.Class, judgeInput),
			Temp:     0,
			MaxToks:  1500,
			JSONMode: true,
		})
		if err != nil {
			r.Logger.Warn("judge failed", "task", t.ID, "err", err)
			continue
		}
		var v Verdict
		if err := json.Unmarshal([]byte(raw), &v); err != nil {
			r.Logger.Warn("judge parse failed", "task", t.ID, "err", err)
			continue
		}
		r.Logger.Info("verdict", "task", t.ID, "min_depth", v.MinViableDepthPct)
	}
	return nil
}

// SyntheticCorpus is a placeholder. Real implementation pulls from
// public-domain corpora (HumanEval, GSM8K, MS-MARCO).
func SyntheticCorpus(domain string, n int) []Task {
	out := make([]Task, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, Task{
			ID:         fmt.Sprintf("%s-%05d", domain, i),
			Class:      "json_classification_short",
			Domain:     domain,
			UserPrompt: "Classify the following review as positive/negative/neutral. Respond with JSON {\"label\": ..., \"confidence\": ...}.\n\nReview: \"The product was okay, nothing special.\"",
		})
	}
	return out
}
