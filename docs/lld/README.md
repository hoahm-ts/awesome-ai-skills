# Low-Level Design Documents

This directory contains Low-Level Design (LLD) documents for this project.
Each LLD captures the internal implementation details of a component or module — package structure, type and interface design, sequence flows, error handling, and testing strategy.

To create a new LLD, copy [`LLD-00-template.md`](LLD-00-template.md), rename it following the `LLD-XX-<short-title>.md` convention, and fill in every section.

> **LLD vs HLD:** An HLD describes the *system-level design* — components, data flows, and API contracts from a bird's-eye view. An LLD describes the *internal implementation* of a single component or module — packages, types, interfaces, and function-level behaviour. An LLD should reference its parent HLD.

---

## Summary

| LLD | Title | Status | Author(s) | Created On | Last Updated |
|-----|-------|--------|-----------|------------|--------------|
| [LLD-00](LLD-00-template.md) | Template | `template` | — | — | — |

---

## Changelog

| Date | LLD | Change |
|------|-----|--------|
| — | [LLD-00](LLD-00-template.md) | Initial template added |

---

## Status Legend

| Status | Description |
|--------|-------------|
| `draft` | Work in progress — not yet reviewed or approved |
| `review` | Under review — awaiting feedback or sign-off |
| `approved` | Approved and ready for implementation |
| `implemented` | Design has been fully implemented |
| `deprecated` | No longer reflects the current implementation |
| `superseded` | Replaced by a newer LLD (reference the superseding LLD) |
| `template` | Placeholder — not a real design document |
