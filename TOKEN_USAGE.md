# Token Usage — MiMo Token Pathology Lab

## Per-task lab cycle (depth prober + judge)

For each task class probed, MiMo is called 6 times: 5 depth-graded runs
plus a single judge.

| Stage              | Agent          | Input tok | Reasoning tok | Output tok | Total/call |
|--------------------|----------------|-----------|---------------|------------|------------|
| 4a. Probe (10%)    | depth-prober   | 2,000     | 1,500         | 800        | 4,300      |
| 4b. Probe (25%)    | depth-prober   | 2,000     | 4,500         | 1,200      | 7,700      |
| 4c. Probe (50%)    | depth-prober   | 2,000     | 9,000         | 1,800      | 12,800     |
| 4d. Probe (75%)    | depth-prober   | 2,000     | 13,500        | 2,200      | 17,700     |
| 4e. Probe (100%)   | depth-prober   | 2,000     | 18,000        | 2,500      | 22,500     |
| 5. Quality Judge   | quality-judge  | 8,000     | 6,000         | 2,000      | 16,000     |
| **Per task**       |                |           |               |            | **81,000** |

Average per probe ≈ 13,000 tokens; the judge is uniformly ~16,000.

## Diagnostic side (per record)

| Stage              | Agent              | Input tok | Reasoning tok | Output tok | Total/call |
|--------------------|--------------------|-----------|---------------|------------|------------|
| 3. Pattern Classifier | pattern-classifier | 1,500  | 2,500         | 800        | 4,800      |

## Treatment side (per batch)

| Stage              | Agent                  | Input tok | Reasoning tok | Output tok | Total/call |
|--------------------|------------------------|-----------|---------------|------------|------------|
| 6. Root-Cause Synth| root-cause-synthesizer | 5,000     | 8,000         | 2,000      | 15,000     |
| 7. Atlas Publisher | atlas-publisher        | 9,000     | 7,000         | 4,500      | 20,500     |

## Daily aggregation (single domain)

```
Lab side
  Tasks/day                    : 200
  Per-task cost (5 probes + 1 judge) : 81,000 tok
  Lab daily total              : 200 × 81,000 = 16.2M

Diagnostic side
  Telemetry rows/day           : ~2,500 (opt-in)
  Complaints scraped/day       : ~150
  Pattern classifier daily     : 2,650 × 4,800 = 12.7M
  Pattern classifier (sampled) : take 30% sample → 3.8M
  (Sampling avoids classifying every duplicate complaint;
   the 70% rest reuse cached classifications.)

Treatment side
  Synthesis batches/day        : 50 (one per active task class)
  Synthesis daily              : 50 × 15,000 = 750K
  Atlas publisher daily        : 1 × 20,500 = 20.5K
  Tenant briefs (5 tenants)    : 5 × 20,500 = 100K
  Community digest             : 1 × 50,000 = 50K

Sub-total                      : 16.2M + 3.8M + 0.92M = 20.9M

Retry / overhead (~10%)        : 2.1M
Daily baseline (1 domain)      : ~23M tok/day
```

## With multi-domain expansion

Lab corpus runs four domains in parallel: **code, math, ops, vision**.

```
Per-domain lab daily : 16.2M
× 4 domains          : 64.8M

But not every domain runs full 200 tasks/day; ramp schedule:

  Week 1   : code only       (16M/day)
  Week 2   : code + math      (32M/day)
  Week 3   : + ops            (48M/day)
  Week 4+  : + vision         (64M/day)

Steady-state monthly average ≈ 40M/day.
```

Adding diagnostic + treatment side: **~45M tok/day** in steady state with
all four domains active.

## Monthly burn projection

```
Month 1 (ramp)         : ~700M tokens
Month 2 (steady-state) : 1.2B-1.4B tokens
```

This is **purpose-built for the 1.6B grant tier**, with token usage that
naturally scales as new domains are added.

## Capacity scaling levers

| Lever                                       | Token impact         |
|---------------------------------------------|----------------------|
| +1 domain (code/math/ops/vision)            | +16M / day per domain|
| 2× tasks/day per domain (200 → 400)         | 2× lab side          |
| 2× depth resolution (5 → 10 levels)         | 2× lab side          |
| Cross-tenant correlation studies            | +5M / day            |
| Real-time scrape (vs every 6h)              | +20% diagnostic      |
| Quality-judge deep-mode (judge of judges)   | +30% lab side        |

Token usage scales linearly with research depth, domain count, and
synthesis cadence — none of which are externally rate-limited.
