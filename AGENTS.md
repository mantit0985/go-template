# AGENTS.md

This repository is a starter spawned from `mantit0985/go-template` ([`Use this template`](https://github.com/mantit0985/go-template/use-this-template)) and operates under the mantit0985 hub standard. The authoritative standard lives in [`mantit0985/.github` → `STANDARDS.md`](https://github.com/mantit0985/.github/blob/main/STANDARDS.md). This file tells any coding agent how to use this starter and how to maintain the repository it spawns.

## What this starter provides

- A minimal Go service skeleton under `src/` with smoke tests under `tests/`
- A `Dockerfile` for the container runtime build consumed by CD
- `.github/workflows/ci.yml` and `.github/workflows/cd.yml` — thin wrappers calling the hub reusable workflows (`mantit0985/.github`) at `@main`; CI runs on every `main` push, CD follows only on CI success

## Using the starter

1. Replace every `[Project Name]` placeholder in `README.md` with the project's name (title pattern: `<Name> - Service Starter Guide`).
2. Implement the service under `src/` and keep the smoke stage healthy; extend tests under `tests/`.
3. Run the toolchain commands below locally before pushing — CI runs the same stages.
4. After cloning run `make hooks` to install the commit-msg guard.

## Toolchain & commands

- go mod download
- go run src/main.go
- go test ./...
- golangci-lint run
- docker build -t go-template .
- docker run -p 8080:8080 --env-file .env go-template

## Versioning

- VERSION at the repository root is the single version source, bumped alongside functional changes.

## Baseline maintenance rituals

1. **Never edit `.github/workflows/ci.yml` or `cd.yml`.** They call the hub reusable workflows at `@main`; behavior evolves in `mantit0985/.github`, proven canary-first in the hub's canary before reaching spawned repos. Local workflow edits break the hub contract.
2. **Symmetry.** Every change lands as exactly one CI run and one CD run per trigger on `main`; CD is gated on CI success.
3. **Green-only `main`.** Fix forward — never commit over a red CI or CD; a failed CD is diagnosed and re-landed green.
4. **README hygiene.** Badges must map to real workflows; keep the single badge row in hub order (CI, CD, Language).
5. **LICENSE.** This starter ships without one; add a LICENSE file when this repository stands up as its own project — the License badge already points at it.
6. **Escalate infrastructure, don't patch around it.** Runner, workflow, or hub issues are fixed in `mantit0985/.github` (or its canary), never worked around locally.
7. **Hook guard.** After cloning run `make hooks` to install the commit-msg guard; commit messages are guarded by the hub guard; AI attribution is forbidden.
