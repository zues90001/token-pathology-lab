# Token Waste Taxonomy

> Living document. Updated as the lab observes new patterns in
> production traffic. Each category links to a sample atlas entry.

## 1. Excessive Self-Reflection

The model wraps reasoning in repeated meta-commentary like
"Let me think...", "Actually, let me reconsider...", inflating output
tokens without adding correctness.

- **Detection:** ratio of meta-phrase tokens / total reasoning tokens > 0.18
- **Mitigation:** suffix prompt with `Be direct. Do not narrate your reasoning steps unless the user explicitly asks.`
- **Severity:** medium (cost), low (correctness impact)

## 2. JSON Wrapping Redundancy

For schema-constrained outputs, models repeat the schema fields in the
reasoning trace before re-emitting them in the final JSON, doubling cost.

- **Detection:** key names appear in both pre-output reasoning and final JSON
- **Mitigation:** use `json_mode=true` + system prompt `Output strict JSON only. No prose, no schema repetition.`
- **Severity:** high (cost), zero (correctness)

## 3. Reasoning Leak into User Output

Reasoning tokens leak into the visible response, mixing
chain-of-thought with the answer. Especially common when system prompts
contain "think step by step".

- **Detection:** scratchpad delimiters (`<thinking>`, `Step 1:`) appear in
  user-facing output
- **Mitigation:** explicit system prompt: `Reasoning is private. Only emit the final answer.`
- **Severity:** medium (cost), high (UX)

## 4. Over-Cautious Refusal Inflation

The model produces 3-5 paragraphs of caveats before answering simple
factual questions. Inflates output without changing correctness.

- **Detection:** > 30% of output tokens precede the actual answer
- **Mitigation:** custom system prompt overriding default refusal
  preamble; or pin to lighter alignment-trained variant
- **Severity:** medium (cost), low (correctness)

## 5. Repeated Chain-of-Thought Cycles

Model retries the same reasoning chain 2-3x within one turn, each time
slightly rephrased. Most visible on long-context inputs.

- **Detection:** n-gram repetition between reasoning blocks > 0.4
- **Mitigation:** lower temperature; explicitly instruct
  `Reason once and commit.`
- **Severity:** high (cost), low (correctness)

## 6. Unnecessary Tool Reasoning Pre-Call

When a tool will be called, models often verbalize the entire tool spec
before deciding which one to use, wasting tokens that the tool layer
already encodes.

- **Detection:** function-call tokens duplicate registered tool schema
- **Mitigation:** strip tool descriptions from reasoning context once
  routed; rely on function-call grammar
- **Severity:** medium-high (cost)

---

## Contributing a Pattern

Open a PR adding a new section + sample atlas entry under
`atlas/<category-slug>/<reproducer-id>.json`.

The publisher CLI validates the entry against the schema described in
`internal/treatment/publisher.go`.
