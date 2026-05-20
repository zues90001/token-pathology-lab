# Privacy

The Pathology Lab handles three different data streams. Each has its own
privacy model.

## 1. Opt-in tenant telemetry

- **Opt-in only** — telemetry endpoint requires an explicit `X-Optin: 1`
  header during onboarding.
- **Anonymised at source** — tenants' SDK strips identifiers before send.
  Prompt content is **never** transmitted in raw form; the SDK sends:
  - SHA-256 prompt hash
  - token counts (prompt / completion / reasoning / total)
  - structural metadata (model, status, latency, retry count, tool count)
  - declared task class (optional, tenant-provided)
- **At-rest encryption** — SQLite ledger encrypted with a tenant-derived
  key when `LAB_ENCRYPT=1`.
- **Retention** — 90 days by default, configurable.
- **Right to delete** — `DELETE /telemetry/tenant/<id>` purges all rows
  on demand.

## 2. Public complaint scrape

- **Public posts only** — Discord scrape requires an opt-in webhook
  granted by MiMo team; forum scrape only ingests posts whose ToS
  permit research use.
- **Attribution removed at ingest** — author handles, avatars, and
  signature blocks are stripped before storage.
- **Aggregated quotes only** — published deliverables never include
  verbatim user posts; only paraphrased patterns.

## 3. Synthetic lab corpora

- **Fully synthetic** — corpora are programmatically generated or sourced
  from public-domain datasets (HumanEval, GSM8K, MS-MARCO).
- **No tenant data** ever enters the lab side. Diagnostic and lab tables
  are physically separate; lab tables are read-only from the diagnostic
  side.

## Published artefacts

| Artefact          | Audience          | Content rule                     |
|-------------------|-------------------|----------------------------------|
| Atlas CSV (CC0)   | MiMo team + public| Aggregate stats only             |
| MiMo team brief   | MiMo team only    | Aggregates + lab examples        |
| Tenant tip        | Single tenant     | Hashed-prompt-id references only |
| Community digest  | Public            | Anonymised aggregates            |

Tenant tips never include another tenant's data.

## Threat model

- **Honest-but-curious analyser operator** — operator can read the
  ledger; cannot recover prompts (only hashes stored).
- **Compromised host** — `LAB_ENCRYPT=1` raises the bar; rotation
  scripts in `internal/storage/rotate.go` re-key on a schedule.
- **Re-identification via task class** — task-class taxonomy is
  intentionally coarse (≤ 50 classes) to prevent fingerprinting.

## Compliance posture

- No PII at rest by default
- No special-category data (health, financial, biometric) accepted at all
- The atlas is research output, not behavioural advertising — no
  third-party trackers, no cross-site IDs, no fingerprinting cookies
