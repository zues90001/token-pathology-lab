# MiMo Team Brief — 2026-05-19

## TL;DR

The largest token-waste cluster this week comes from
**`json_classification_short`**: short-form classification tasks where
the model burns 4-5× more reasoning tokens than necessary to preserve
output quality.

- Real-world signal: 1,450 telemetry rows + 31 unique forum complaints
  in the past 7 days
- Lab finding: depth-pct 25 preserves quality (0.95 score) vs depth-pct
  100 (0.97). Below 25, output schema breaks (1 of 3 required JSON
  fields drops).
- Estimated ecosystem savings if remediation lands by default: **60M
  tokens/month** across observed tenants.

## Recommended training adjustment

Bias the model toward shallower reasoning when the user prompt declares a
strict short-form output schema (single label, boolean, ranked list of
< 5 items). Concretely: detect schema declarations in the prompt and
pre-cap reasoning at depth-pct ≤ 25 unless the model itself flags the
task as ambiguous.

This can land as either:

1. A **training-data signal**: include the atlas dataset in the next
   fine-tune so the model learns the depth-vs-quality plateau directly.
2. A **system prompt heuristic** in the base hosted MiMo to short-circuit
   reasoning when the schema is short.

## Supporting evidence

| Source            | Count | Severity (avg) | Top failure mode                            |
|-------------------|-------|----------------|---------------------------------------------|
| Tenant telemetry  | 1,450 | 0.74           | over-reasoning on 1-field classification    |
| Discord scrape    | 21    | 0.69           | "why did this prompt cost so much?"         |
| Forum scrape      | 10    | 0.71           | thread comparing MiMo vs other LLM costs    |

Lab quality curve (0.0 worst, 1.0 best):

```
depth_pct  10   25   50   75   100
quality   .62  .95  .96  .96  .97
```

Plateau begins at 25% — extra depth adds verbosity, not correctness.

## Proposed atlas row (CC0)

See `atlas/2026-05-19/json_classification_short.json`. CSV roll-up
appended to the public dataset at midnight UTC.

## Counterfactual

If remediation lands across the observed tenants:

- Tokens saved: 14,800 per call × 4,000 calls/day across tenants
- Daily savings: ~60M tokens
- Monthly savings: ~1.8B tokens

## Open questions for the MiMo team

1. Is the schema-detection heuristic feasible inside the base hosted
   model, or does it require fine-tune?
2. Are there task classes we should *not* shorten even when the schema
   looks short (e.g. legal classifications where reasoning matters even
   for binary verdicts)?
3. Should the public atlas include per-tenant savings projections, or
   only aggregate?

We'd love feedback before publishing.
