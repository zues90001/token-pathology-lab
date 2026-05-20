# MiMo Token Pathology Lab

> Bridge user complaints → controlled lab experiments → open dataset that
> MiMo team can ship as training data. Diagnose where the token waste comes
> from, prove the root cause, and publish the fix.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
![Status](https://img.shields.io/badge/status-alpha-orange)
![Stack](https://img.shields.io/badge/lang-Go-00ADD8)

## Why this exists

The MiMo Discord and forums surface a recurring complaint: **token usage
feels disproportionate to prompt complexity**. Users see surprising bills
on short prompts and have no way to find out why.

Existing community workarounds (proxies, dashboards) treat the symptom.
They tell the user "your spend looks high" but don't trace the cause back
to the model itself. The MiMo team gets no actionable signal either —
just complaint threads.

This project closes the loop in three movements:

1. **Diagnostic** — collect complaint signal (opt-in user telemetry +
   public Discord/forum scrape) and classify the waste pattern.
2. **Lab** — run controlled experiments on synthetic task corpora across
   five reasoning-depth levels, measure where output quality plateaus.
3. **Treatment** — publish three deliverables daily:
   - **Open atlas dataset** (training-ready CSV) for the MiMo team
   - **Tenant remediation tips** for opt-in users
   - **Anonymised community digest** for everyone else

The resulting Reasoning Atlas is the first open dataset that lets the
MiMo team train against measured user pain.

## Why MiMo?

Every reasoning step in the pipeline calls MiMo:

- **Pattern classifier** — read complaint + telemetry, classify waste
- **Depth prober** — run task at 5 reasoning-depth levels (5× MiMo calls
  per task)
- **Quality judge** — compare 5 outputs, find the plateau point
- **Root-cause synthesizer** — match real-world complaint to lab finding
- **Atlas publisher** — write daily MiMo team brief + tenant tips

Cheaper instruct models can't preserve reasoning depth comparisons
faithfully — they collapse the very signal we're trying to measure.

## Architecture (one-paragraph)

7-stage pipeline split across three sides. The **diagnostic side**
collects signal from opt-in tenants (`telemetry-collector`, deterministic)
and public forums (`complaint-aggregator`, deterministic), then
`pattern-classifier` (MiMo) tags real-world waste. The **lab side** runs
synthetic task corpora through `depth-prober` (MiMo, 5× per task) and
`quality-judge` (MiMo) to find the minimum viable reasoning depth per
task class. The **treatment side** uses `root-cause-synthesizer` (MiMo)
to match real-world patterns to lab findings, then `atlas-publisher`
(MiMo) ships three deliverables. Full diagram in
[`ARCHITECTURE.md`](ARCHITECTURE.md).

## Token economics

- **29M tokens/day** baseline (single domain)
- 5 MiMo stages, but the depth-prober is the workhorse (5× per task)
- Full breakdown: [`TOKEN_USAGE.md`](TOKEN_USAGE.md)
- Burn projection: **870M-1.2B tokens/month** (designed for the 1.6B
  grant tier)

## Repo layout

```
token-pathology-lab/
├── README.md
├── ARCHITECTURE.md
├── PROMPTS.md
├── TOKEN_USAGE.md
├── GRANT_PITCH.md
├── PRIVACY.md
├── LICENSE
├── go.mod
├── cmd/
│   ├── collector/main.go     ← opt-in telemetry collector
│   ├── prober/main.go        ← lab experiment runner
│   └── publisher/main.go     ← daily atlas + briefs publisher
├── internal/
│   ├── diagnostics/          ← telemetry + complaint aggregation
│   ├── lab/                  ← depth prober + quality judge
│   ├── treatment/            ← root-cause synthesizer + atlas publisher
│   ├── mimo/                 ← shared MiMo client
│   └── storage/              ← SQLite ledger
└── examples/
    ├── sample_atlas_entry.json
    ├── sample_mimo_team_brief.md
    └── sample_tenant_tip.md
```

## Quick start

```bash
git clone https://github.com/<account>/token-pathology-lab
cd token-pathology-lab

go build ./cmd/...

export MIMO_API_KEY="..."

./bin/collector --listen :8090       # opt-in telemetry endpoint
./bin/prober --domain code --tasks 200
./bin/publisher --output ./atlas/
```

## Privacy

See [`PRIVACY.md`](PRIVACY.md). Short version:

- Telemetry: **opt-in only**, anonymised at source, encrypted at rest
- Complaint scraping: **public posts only**, attribution removed before
  storage
- Lab corpora: **fully synthetic**, no user data ever
- All published artefacts: **aggregates only**, never prompt verbatim

## License

MIT. The atlas dataset is released under CC0 so the MiMo team can adopt
it directly into training pipelines without license friction.
