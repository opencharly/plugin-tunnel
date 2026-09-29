# AGENTS.md — plugin-tunnel

Standalone plugin repo for the tunnel execution leg (`verb:tunnel`). The plugin
is a Go module at `candy/plugin-tunnel/` (module path
`github.com/opencharly/plugin-tunnel/candy/plugin-tunnel`); the root `charly.yml`
only declares `discover: candy` so the repo is a project and its candy is
scanned.

Canonical files:

- `candy/plugin-tunnel/charly.yml` — the `plugin-tunnel:` candy entity
  (`plugin:` block, `plan:` checks).
- `candy/plugin-tunnel/plugin.go` — the provider (`NewProvider`/`NewMeta`) +
  the `start`/`stop`/`setup`/`plan` methods.
- `candy/plugin-tunnel/tunnel_exec.go` — the tailscale/cloudflared execution leg.
- `candy/plugin-tunnel/schema/tunnel.cue` — the self-contained input schema.
- `.github/workflows/tag-on-merge.yml` — CalVer tag + `CHANGELOG/` on merge.
- `README.md` — user overview only; never agent guidance.

## Load these skills first (R0)

- `/charly-internals:plugin` — the plugin authoring reference: the `plugin:`
  block, the unified Provider model, the per-plugin CUE-schema contract,
  placement. Load before touching the provider or schema.
- `/charly-core:deploy` — the deploy surface whose tunnel lifecycle this verb
  executes.
- `/charly-internals:git-workflow` — before any git/PR action.

## Build / validate / test

- `go build ./...` in `candy/plugin-tunnel/` — compile the plugin module.
- `go test ./...` in `candy/plugin-tunnel/` — the plugin's Go tests
  (`tunnel_exec_test.go`).
- `charly box validate` at the repo root — the structural check (the candy +
  `plugin:` block, CUE schema).
- The merge gate is the **org-wide** `charly/pr-validator` (required check
  `validate / validate`, defined in `opencharly/.github`); this repo has **no**
  per-repo candy gate.
- The `plan` dry-run method is the creds-free evidence; the live
  `check-tunnel-pod` bed in `box/fedora` asserts the built argv.

## Modify this repo

- Edit the `plugin-tunnel:` candy entity, the Go source, and
  `schema/tunnel.cue` **together** — the schema is the single source for the
  `params/` struct, so a field change not mirrored in the schema desyncs the
  generated types.
- The provider is **dual-placement** (compiled-in default, or out-of-process);
  do not describe it as one or the other.
- The resolution half lives in `sdk/deploykit/tunnel_resolve.go`; only the
  execution leg belongs here.

## Landing

- PR-only. Every change lands through a pull request; the org-required
  `charly/pr-validator` validates the diff and body and arms native auto-merge on
  PASS. Direct pushes to `main` are blocked.
- History lives in `CHANGELOG/` (written by `tag-on-merge` at merge time); the PR
  body IS the changelog.
- The authoritative rulebook is the umbrella `AGENTS.md` in
  `opencharly/opencharly` and `charly/AGENTS.md` in the charly repo. Do not
  restate its rules here.
