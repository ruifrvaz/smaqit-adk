---
status: In Progress
created: "2026-08-23"
mode: Assisted
started: "2026-08-23"
---

# Migrate smaqit-adk to daisy-tribe Org (Private Repo)

## Description

Move smaqit-adk from `github.com/ruifrvaz/smaqit-adk` to a new private repository under the
`daisy-tribe` GitHub org. The repo move itself (new remote, new scoped PAT) is straightforward and
handled outside this task. What this task covers: rewriting every repo-identity reference in source
and docs, and — the real risk — repairing `install.sh`'s distribution flow, which currently assumes
unauthenticated public GitHub access on two separate calls (`api.github.com/repos/<repo>/releases`
and a plain `github.com/<repo>/releases/download/...` binary URL). Both 404 unauthenticated against a
private repo, and so would a bare `curl raw.githubusercontent.com/.../install.sh` fetch of the script
itself.

The resolved design: `gh` CLI becomes the sole, required install-time dependency — no curl-based
fallback. `install.sh` is attached as a version-pinned release asset (via `post-merge-release.yml`,
alongside the existing platform binaries) rather than fetched live off `main`, so script logic and
binary/API expectations never drift apart. The end-user install experience stays a single
copy-paste-into-bash command, now two chained `gh`/bash calls instead of one `curl | bash`:

```
gh release download --repo daisy-tribe/smaqit-adk --pattern install.sh -D /tmp/smaqit-adk-install --clobber && bash /tmp/smaqit-adk-install/install.sh
```

`install.sh` itself then uses `gh release download` again internally to fetch the correct platform
binary — same tool, same auth (interactive `gh auth login`, or `GH_TOKEN`/`GITHUB_TOKEN` env var for
headless/prod use), no raw-content dependency anywhere in the chain. Exact `gh release download` flags
should be verified against `gh release download --help` during implementation rather than assumed from
this description.

There is no self-update/upgrade mechanism anywhere in the codebase today (confirmed via discovery), so
there is nothing to repair on that front. The CI/CD release pipeline (`post-merge-release.yml`,
`test-integration.yml`) is unaffected by the visibility change itself — both already authenticate via
the Actions-provided `GITHUB_TOKEN` and `actions/checkout@v4` — except for the one deliberate addition
of attaching `install.sh` as a release asset.

## Issue Triage Context

**Mode:** Skip
**Technologies:** Bash (install.sh), Go (go.mod/module paths), GitHub Actions, GitHub CLI (gh)
**Platforms/Environments:** GitHub (private org repo daisy-tribe/smaqit-adk)
**Features/Integrations:** install.sh distribution, gh release download, CI/CD release workflow
**Versions/Constraints:** Current adk version 2.0.1; migrating from github.com/ruifrvaz/smaqit-adk

## Design Decisions

- **Repo visibility: private.** Confirmed by user; this is the entire reason install.sh needs rework.
- **`gh` CLI is the sole required install-time dependency.** No curl-based fallback is kept — the
  existing zero-dependency curl-only install.sh is deliberately replaced, not extended. Discussed and
  rejected: (a) git-clone + local `make build` (a dev workflow, not a prod release install — explicitly
  rejected by the user), (b) hand-rolled curl + GitHub API asset-endpoint parsing in bash (works, but
  `gh release download` already wraps exactly this — resolving the release, matching the asset by name,
  authenticating, downloading — so hand-rolling it duplicates logic for no benefit).
- **`install.sh` is version-pinned as a release asset**, attached to each release via
  `post-merge-release.yml`, not fetched live off `main` via `gh api repos/.../contents/install.sh`.
  Keeps the installer script's expectations in lockstep with the binary/API shape of the release it
  installs.
- **Install UX stays a single copy-paste one-liner** — now `gh release download ... && bash ...`
  instead of `curl ... | bash`, chained with `&&` so it remains one line to copy.
- **No self-update mechanism exists today** — confirmed via repo-wide grep (only `install.sh` and
  unrelated bench event-format strings reference "update"/"releases"). Out of scope; nothing to fix.
- **`CHANGELOG.md`'s ~18 historical compare-URLs are left pointing at the old org**, treated as an
  accurate historical record rather than rewritten.

## Implementation Steps

1. **Rewrite repo-identity references** (after the repo migration itself is done):
   - `src/go.mod:1`, `installer/go.mod:1,5,9`, `tests/go.mod:1` — module path
     `github.com/ruifrvaz/smaqit-adk/...` → `github.com/daisy-tribe/smaqit-adk/...`
   - `installer/main.go:11` and `src/benchcli/bench.go:16` — matching import statements
   - `installer/main.go:83,93` — help-text URLs
   - `README.md:31,36,44,156`, `docs/wiki/extending-smaqit.md:29` — product link, clone URL, attribution
   - `install.sh:3,11-13,24` — usage-comment URLs and the `REPO=` variable
2. **Attach `install.sh` as a release asset** in `.github/workflows/post-merge-release.yml`'s existing
   `softprops/action-gh-release@v2` step — add it to the `files:` list alongside the platform binaries
   already published there.
3. **Rewrite `install.sh`'s fetch logic to be `gh`-only:**
   - Replace the `api.github.com/repos/${REPO}/releases` listing call with `gh release download`'s
     built-in release resolution (defaults to latest when no tag is given).
   - Replace the plain `github.com/${REPO}/releases/download/...` binary download with a
     `gh release download --pattern <binary-name-for-this-OS/ARCH>` call.
   - Add a preflight check: fail fast with a clear error message if `gh` is not installed or
     `gh auth status` reports not authenticated (and no `GH_TOKEN`/`GITHUB_TOKEN` is set — `gh` reads
     these natively, no custom token handling needed).
   - Remove the now-unused `curl`-based fetch code entirely — no fallback path.
4. **Replace the README/wiki install one-liner** in `README.md` and `docs/wiki/extending-smaqit.md`
   with the new `gh`-based, version-pinned, single-copy-paste command (see Description for the exact
   shape).
5. **End-to-end verification** — from a clean environment, run the real new one-liner against the real
   `daisy-tribe/smaqit-adk` private repo, both with an interactive `gh auth login` session and
   separately with just `GH_TOKEN` set.

## Known Issues Triage

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [ ] Repo pushed to `github.com/daisy-tribe/smaqit-adk` with new remote and scoped PAT configured
- [ ] All 3 `go.mod` module paths and both Go import statements updated to `daisy-tribe/smaqit-adk`;
      `make build` passes cleanly
- [ ] `post-merge-release.yml` attaches `install.sh` as a release asset alongside platform binaries
- [ ] `install.sh` fetches its release binary exclusively via `gh release download`, with no
      curl/raw.githubusercontent.com calls remaining; fails fast with a clear error when `gh` is
      missing or unauthenticated
- [ ] `README.md` and `docs/wiki/extending-smaqit.md`'s install one-liner replaced with the new
      `gh`-based, version-pinned, single-copy-paste command
- [ ] The real new one-liner exercised end-to-end against the real private repo, both with interactive
      `gh auth login` and with `GH_TOKEN` set, confirmed working

## Findings

[Populated by smaqit.task-complete. Do not fill in manually before task is complete.]

**Implementation approach:**
- TBD

**Decisions made:**
- TBD

**Blockers encountered:**
- TBD

**Follow-up identified:**
- TBD

## Files to Create / Modify

| File | Action |
|------|--------|
| `install.sh` | Modify — replace curl-based fetch with `gh release download`-only flow, add auth preflight check |
| `.github/workflows/post-merge-release.yml` | Modify — attach `install.sh` as a release asset |
| `README.md` | Modify — repo-identity references + new install one-liner |
| `docs/wiki/extending-smaqit.md` | Modify — repo-identity references + new install one-liner |
| `installer/main.go` | Modify — import path, help-text URLs |
| `src/benchcli/bench.go` | Modify — import path |
| `src/go.mod` | Modify — module path |
| `installer/go.mod` | Modify — module/require/replace paths |
| `tests/go.mod` | Modify — module path |

## Notes

- Planned via `smaqit.task-plan` (Mode A) in this session, 2026-08-23. Discovery (2 Explore agents)
  confirmed: no self-update mechanism exists anywhere in the codebase; the CI/CD release pipeline is
  unaffected by repo visibility (uses `actions/checkout@v4` + Actions-provided `GITHUB_TOKEN`); the
  Go-import blast radius is small (2 import statements, 3 `go.mod` files); 41 total "ruifrvaz" string
  hits across 10 files, none inside `.github/workflows/`, `framework/`, `templates/`, or `agents/`.
- Design went through two course corrections during planning, both driven directly by user pushback:
  git-clone + local `make build` was proposed first and explicitly rejected ("that's for devs... we
  want a prod (release) installation"); a hand-rolled curl + GitHub API asset-endpoint approach was
  proposed second and superseded once the user pointed out `gh release download` already wraps that
  exact endpoint.
- Open, non-blocking considerations surfaced during planning (not acted on, left for whoever executes):
  confirm the new scoped PAT (or `gh auth login` session) has at least `Contents:read` on the new repo;
  whether the org migration is a GitHub "Transfer repository" (redirects old URLs) or a fresh push to a
  new repo (old links just 404) doesn't change this task's implementation either way.
