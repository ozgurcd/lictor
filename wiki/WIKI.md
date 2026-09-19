# Lictor wiki contract

This repository owns its wiki. Read index.md, agent-rules.md and repos/lictor.md.
PROJECT_DESC.md is owner-controlled; PROJECT_SPEC.md describes implemented code.
Every committed slice appends a co-versioned ledger row in repos/lictor.md and
an entry in log.md. Never fall back to a parent or sibling wiki.
Run `make wiki-check`; it selects this wiki explicitly with Achta. Mechanical
freshness is not proof of prose accuracy; review source before updating facts.
