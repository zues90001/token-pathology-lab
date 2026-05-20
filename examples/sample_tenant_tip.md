# Tenant remediation tip — 2026-05-19

**Tenant:** demo-tenant
**Detected pattern:** `json_classification_short` (12 prompts in the past
24h)

## What we found

12 of your prompts in the past 24h match the `json_classification_short`
task class. They produced ~22,500 tokens each, with reasoning tokens
dominating (~18,000 / call).

Lab data shows quality plateaus at 25% reasoning depth for this task
class. You're paying for 4× more reasoning than needed.

## Suggested remediation

Prepend this prefix to your system prompt:

> *Answer concisely without intermediate analysis. Skip the rationale
> unless explicitly asked.*

## Expected savings

- Per-call savings: **~14,800 tokens** (≈ 65%)
- Daily savings (12 calls): **~178,000 tokens**
- Monthly savings (extrapolated): **~5.3M tokens**

## How to validate

The `verify` endpoint will re-run one of your prompts under both the
original and remediated system prompt, and report whether the output
schema and core decision match. No live calls are mutated; this is a
parallel test.

```bash
curl -X POST https://lab.example.com/verify \
  -d '{"tenant_id": "demo-tenant", "task_id": "abc123"}'
```

If the verifier returns `equivalent: true`, you can adopt the prefix
across your prompts. If it returns `partial`, review the diff before
adopting.

## Privacy reminder

This tip is generated from telemetry you opted into. We never store
your prompt text — only hashes and structural metadata. To opt out at
any time, send `DELETE /telemetry/tenant/demo-tenant`.
