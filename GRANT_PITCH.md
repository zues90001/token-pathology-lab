# Grant Pitch — MiMo Token Pathology Lab

> Direct copy-paste content for the MiMo 100T application form.

---

## Section 1 — Project name + tagline
**MiMo Token Pathology Lab** — Bridge user complaints to controlled lab
experiments to a training-ready open atlas. Diagnose where the token
waste comes from, prove the root cause, ship the fix back to the
ecosystem.

---

## Section 2 — Problem statement
The MiMo Discord and forum surface a recurring complaint: token usage
feels disproportionate to prompt complexity. Users see surprising bills
on short prompts and can't trace the cause. The MiMo team receives
hundreds of complaint threads each month with no actionable signal — no
labels, no controlled experiments, no training data.

Existing community workarounds (proxies, dashboards) treat the symptom on
the user side. None of them feed root-cause data back to the model team.
There is no public dataset that maps "task class → minimum viable
reasoning depth → expected savings". Without that dataset, training next
versions of MiMo to be more efficient is guesswork.

This project is the missing research layer that closes the loop.

---

## Section 3 — Solution architecture
A 7-stage pipeline split across three sides:

```
DIAGNOSTIC SIDE  →  LAB SIDE  →  TREATMENT SIDE
   (real signal)    (controlled)    (deliverables)
```

**Diagnostic** — `telemetry-collector` (det.) ingests opt-in tenant
telemetry; `complaint-aggregator` (det.) scrapes public Discord/forum
threads; `pattern-classifier` (MiMo) tags each record by task class and
waste profile.

**Lab** — `depth-prober` (MiMo, 5× per task) runs the same prompt at
five reasoning-depth levels (10/25/50/75/100%); `quality-judge` (MiMo)
compares the five outputs and emits the minimum viable depth.

**Treatment** — `root-cause-synthesizer` (MiMo) joins real-world
patterns to lab findings; `atlas-publisher` (MiMo) ships four daily
artefacts: open atlas CSV (CC0, training-ready), MiMo team brief,
per-tenant remediation tip, anonymised community digest.

5 of 7 stages call MiMo. The depth-prober is the workhorse: 200 tasks/
day × 5 probes = 1,000 MiMo calls/day from that single agent alone.

---

## Section 4 — Why MiMo
Every reasoning step is comparative ("does 25% depth preserve quality vs
50%?"). Comparing reasoning chains is exactly the long-chain task MiMo
was designed for. Cheaper instruct models collapse the comparison,
emitting "all five outputs look the same" and destroying the signal.

Beyond capability, this project's whole reason to exist is to **improve
MiMo** — using MiMo to study MiMo, then giving the team a CC0 dataset
they can fold into training. There's no other model where this story
makes sense.

---

## Section 5 — Token usage table

| Stage                  | Tokens/call | Calls/day | Tokens/day |
|------------------------|-------------|-----------|------------|
| Pattern classifier     | 4,800       | 800       | 3.8M       |
| Depth prober (5/task)  | 13,000 avg  | 1,000     | 13.0M      |
| Quality judge          | 16,000      | 200       | 3.2M       |
| Root-cause synthesizer | 15,000      | 50        | 0.75M      |
| Atlas publisher        | 20,500      | 1         | 20.5K      |
| Tenant briefs          | 20,500      | 5         | 0.10M      |
| Community digest       | 50,000      | 1         | 50K        |
| Retry / overhead       | -           | -         | 2.1M       |
| **Single domain**      |             |           | **~23M**   |
| × 4 domains (steady)   |             |           | **~45M**   |

**Monthly burn: 1.2–1.4B tokens** at steady state. Designed for the 1.6B
grant tier. Token usage scales linearly with research depth and domain
count — never externally rate-limited.

---

## Section 6 — Continuous operation
- Telemetry endpoint: 24/7 listener (opt-in tenants)
- Complaint scrape: every 6 hours
- Lab prober: 200 tasks/day per active domain
- Atlas publisher: midnight UTC daily
- Community digest: weekly
- Domain ramp: code → math → ops → vision (4-week schedule)

---

## Section 7 — Open-source commitment
- License: **MIT** for code, **CC0** for the atlas dataset
- Repo: `github.com/<account>/token-pathology-lab`
- Roadmap milestones tied to token-budget burn:
  - 200M tokens: code domain, single-tenant pilot
  - 600M tokens: 4 domains, full diagnostic loop, first MiMo team brief
  - 1.4B+ tokens: deep-mode judge, real-time scrape, cross-tenant studies

---

## Section 8 — Team
Solo developer with prior MiMo grant track record (Round 1: mymimo;
Round 2: reasoning-arena, both approved). Strong open-source distribution
muscle: prior projects rank #1 on GitHub trending in their week.

The project's research design pulls directly from Wann's existing
benchmark work — reasoning-arena is essentially the experimental
infrastructure that this lab extends.

---

## Section 9 — Demo / PoC
- GitHub repo with runnable collector + prober + publisher
- Sample atlas entry committed at `examples/sample_atlas_entry.json`
- Sample MiMo team brief at `examples/sample_mimo_team_brief.md`
- Sample tenant remediation tip at `examples/sample_tenant_tip.md`
- 60-second screen capture: prober runs 5-depth experiment on a single
  task, judge picks the plateau, publisher emits the atlas row
