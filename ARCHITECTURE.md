# Architecture — MiMo Token Pathology Lab

## High-level data flow

```
══════════════ DIAGNOSTIC SIDE (real-world signal) ══════════════

  opt-in tenants                          public forums
      │                                          │
      ▼                                          ▼
  ┌────────────────────────┐         ┌────────────────────────┐
  │ 1. Telemetry Collector │         │ 2. Complaint Aggregator│
  │  - opt-in only         │         │  - Discord scrape      │
  │  - anonymise at source │         │  - forum scrape        │
  │  - SQLite              │         │  - dedupe + redact     │
  └──────────┬─────────────┘         └──────────┬─────────────┘
             │                                   │
             └─────────────┬─────────────────────┘
                           ▼
              ┌────────────────────────┐
              │ 3. Pattern Classifier ★│  ← MiMo
              │  - tag waste type      │
              │  - severity score      │
              │  - task class label    │
              └──────────┬─────────────┘
                         │
                         ▼
              [diagnostic_findings table]
                         │
══════════════════════════│════════════════════════════════════════
                          │
══════════════ LAB SIDE (controlled experiments) ══════════════════
                          │
              ┌───────────┴────────────┐
              │ synthetic task corpus  │
              │ (code/math/ops/vision) │
              └───────────┬────────────┘
                          ▼
              ┌────────────────────────┐
              │ 4. Depth Prober ★      │  ← MiMo (5× per task)
              │  - 10/25/50/75/100%    │
              │  - capture each output │
              └──────────┬─────────────┘
                         ▼
              ┌────────────────────────┐
              │ 5. Quality Judge ★     │  ← MiMo
              │  - compare outputs     │
              │  - find quality plateau│
              │  - emit min_viable_pct │
              └──────────┬─────────────┘
                         │
                         ▼
                [lab_findings table]
                          │
══════════════════════════│═════════════════════════════════════════
                          │
══════════════ TREATMENT SIDE (deliverables) ═══════════════════════
                          │
              ┌───────────┴────────────┐
              │ diagnostic + lab join  │
              └───────────┬────────────┘
                          ▼
              ┌────────────────────────┐
              │ 6. Root-Cause Synth ★  │  ← MiMo
              │  - real-world ↔ lab    │
              │  - propose remediation │
              │  - score correlation   │
              └──────────┬─────────────┘
                         ▼
              ┌────────────────────────┐
              │ 7. Atlas Publisher ★   │  ← MiMo
              │  emits 3 artefacts:    │
              │  • MiMo team brief     │
              │  • Tenant remediation  │
              │  • Community digest    │
              │  • Open atlas CSV      │
              └────────────────────────┘
                         │
                         ▼
            atlas/  briefs/  tips/  digests/
```

## Agents

### 1. Telemetry Collector (deterministic)
Receives opt-in tenant telemetry over HTTPS. Strips PII at the moment of
ingest, hashes prompt content, keeps only token counts and structural
metadata. No prompt text reaches storage by default.

**No MiMo.** See `internal/diagnostics/collector.go`.

### 2. Complaint Aggregator (deterministic)
Scrapes the MiMo Discord (with opt-in webhook from MiMo team) and public
forums (e.g. Reddit r/LLM threads where MiMo is named). Deduplicates,
normalises, and removes user attribution before storage.

**No MiMo.** See `internal/diagnostics/aggregator.go`.

### 3. Pattern Classifier (MiMo)
For each diagnostic record (telemetry row or complaint), MiMo classifies:
- **task class** (json_classification_short, code_review_long, ...)
- **waste profile** (verbose-system, over-reasoning, retry-loop, ...)
- **severity score** (0-1)

**MiMo call:** `~1,500 input + 2,500 reasoning + 800 output ≈ 4,800 tok`

### 4. Depth Prober (MiMo, workhorse)
For each synthetic task, MiMo runs the same prompt at five reasoning-
depth caps: 10%, 25%, 50%, 75%, 100% (controlled via system-prompt
hints + temperature lock). All five outputs persist for comparison.

**MiMo call:** `~12,000 tok per probe × 5 probes per task`

### 5. Quality Judge (MiMo)
Reads all five depth-graded outputs side-by-side and emits:

- **min_viable_depth_pct** — the lowest depth where quality preserves
- **quality_curve** — per-depth quality score
- **failure mode** at sub-viable depths

**MiMo call:** `~8,000 tok per task` (single judge call comparing 5)

### 6. Root-Cause Synthesizer (MiMo)
Joins diagnostic findings with lab findings on **task class**, computes
correlation, drafts:

- proposed remediation prompt prefix
- proposed model-side fix (training-data hint)
- estimated ecosystem savings

**MiMo call:** `~15,000 tok per synthesis batch`

### 7. Atlas Publisher (MiMo)
Daily output:

- **MiMo team brief** — root cause + recommended training adjustment
- **Per-tenant remediation tip** — concrete prompt prefix per detected
  task class
- **Community digest** — anonymised aggregate
- **Open atlas CSV** — appended row per (task_class, depth, savings)

**MiMo call:** `~20,500 tok per day` (briefs are short by design)

## Synthetic corpus design

Lab corpora cover four domains and grow over time:

| Domain     | Example task classes                            |
|------------|-------------------------------------------------|
| code       | json_classification_short, code_review_long     |
| math       | arithmetic_step, algebra_proof                  |
| ops        | log_summarise, alert_triage                     |
| vision     | image_caption_short, doc_layout_describe        |

Corpora are 100% synthetic / public-domain. No tenant prompts ever enter
the lab.

## Data lineage and lifecycle

```
Tenant telemetry  ─►  hashed at ingest  ─►  diagnostic_findings table
                                              │
Public complaints ─►  redacted at ingest ─►  └► classified by MiMo
                                              │
Synthetic corpus  ─►  raw + persisted    ─►  lab_findings table
                                              │
                                              ▼
                                       Root-cause synthesis
                                              │
                                              ▼
                          atlas/ + briefs/ + tips/ + digests/
```

## Continuous mode

- Telemetry endpoint: 24/7 listener
- Complaint scrape: every 6 hours
- Depth prober: 200 tasks/day across 4 domains
- Atlas publisher: midnight UTC daily
- Community digest: weekly

## Why long-chain reasoning matters here

Detecting "this task only needs 25% depth to preserve quality" is itself a
reasoning task: the judge has to compare five outputs, identify the
plateau, justify why 25% is sufficient and 10% isn't. Cheaper models
fail this comparison and emit "all five look the same" — destroying the
exact signal we need.
