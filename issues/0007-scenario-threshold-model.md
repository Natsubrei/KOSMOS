---
labels: [wayfinder:grilling]
assignee:
---

# 0007 · Scenario and threshold model

## Question

How do conclusions derive from evidence: hardcoded diagnosis pipelines, a rule/data-driven scenario model, or a hybrid? Where do thresholds live (code, config, database), who can change them, and how does a scenario declare its required sources and degrade audibly when one is missing? Design against the [report model](../.agents/notes/implemented/architecture/2026-10-26-report-evidence-model.md) so every derived Conclusion carries its Evidence. Threshold configuration feeds the Predicate recorded on Evidence; a required source that is missing is a Missing source.
