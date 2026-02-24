# smaqit-adk Handover

## Current Status

**Date:** 2026-02-24
**Phase:** ADK extraction complete — agents cleaned, renamed SDK→ADK, build verified; pending first commit + push + L2 agent cleaning
**Next:** `git commit`, `gh repo create ruifrvaz/smaqit-adk --private --source=. --push`, `git tag adk-v0.1.0`, then clean L2 agent content

## What Was Done

### Repository Split Completed

1. **Created new smaqit-adk repository** at `/home/ruifrvaz/projects/smaqit-adk/`
2. **Copied ADK-relevant files** from smaQit:
   - Framework files (5): SMAQIT.md, AGENTS.md, TEMPLATES.md, ARTIFACTS.md, PROMPTS.md
   - Generic agent templates (3): base-agent, specification-agent, implementation-agent
   - Generic compilation rules (3): base, specification, implementation
   - Level agents (3): L0, L1, L2 (need cleaning)
   - new-agent prompt template
   - Installer files

3. **Created new ADK infrastructure:**
   - Self-contained installer (main.go, Makefile, go.mod)
   - README.md (ADK-focused)
   - CHANGELOG.md (initial 0.1.0 entry)
   - .gitignore
   - install.sh script
   - GitHub workflows (release.yml, test-integration.yml)

### Files Created

```
/home/ruifrvaz/projects/smaqit-adk/
├── README.md                           # ADK documentation
├── CHANGELOG.md                        # Version 0.1.0
├── LICENSE                             # Copied from smaQit
├── .gitignore                          # Build artifacts
├── install.sh                          # Installer script
├── framework/                          # 5 generic principle files
│   ├── SMAQIT.md
│   ├── AGENTS.md
│   ├── TEMPLATES.md
│   ├── ARTIFACTS.md
│   └── PROMPTS.md
├── templates/
│   └── agents/
│       ├── base-agent.template.md
│       ├── specification-agent.template.md
│       ├── implementation-agent.template.md
│       └── compiled/
│           ├── base.rules.md
│           ├── specification.rules.md
│           └── implementation.rules.md
├── agents/                             # Level agents (NEED CLEANING)
│   ├── smaqit.L0.agent.md
│   ├── smaqit.L1.agent.md
│   └── smaqit.L2.agent.md
├── prompts/
│   └── smaqit.new-agent.prompt.md
├── installer/
│   ├── main.go                         # Self-contained
│   ├── Makefile                        # No ../ references
│   └── go.mod                          # github.com/ruifrvaz/smaqit-adk/installer
├── docs/
│   └── wiki/
└── .github/
    └── workflows/
        ├── release.yml                 # ADK release automation
        └── test-integration.yml        # Structure validation
```

## Critical Next Steps

### 1. ✅ Clean Level Agents

**Status:**
- ✅ `agents/smaqit.L0.agent.md` — Cleaned: removed LAYERS.md/PHASES.md input refs, generalized principle examples
- ✅ `agents/smaqit.L1.agent.md` — Cleaned: fixed file inventory, generalized placeholder standards
- ⚠️ `agents/smaqit.L2.agent.md` — NOT YET FULLY CLEANED: body still contains 8 product agents in Output, non-existent layer rules in Input, hardcoded placeholder resolution tables (Business/BUS etc.), smaQit-specific examples. Only the 1-line "ADK extensibility" rename was applied.
- ✅ `templates/agents/compiled/specification.rules.md` — Cleaned: removed base redundancies, generalized layer→domain
- ✅ `templates/agents/compiled/implementation.rules.md` — Cleaned: removed smaQit CLI coupling, restored generic State Tracking

**Remaining L2 work:**
- Remove 8 smaQit product agents from Output section
- Remove 8 non-existent layer/phase `.rules.md` from Input
- Rewrite Compilation Architecture to describe 3-way merge
- Remove hardcoded Placeholder Resolution tables (Business/BUS, Functional/FUN etc.)
- Replace smaQit-specific form examples with generic `[DOMAIN]`/`[PREFIX]` placeholders
- Fix prompt path and remove `.smaqit/logs/` compilation log requirement

### 2. ✅ Test Build

```bash
cd /home/ruifrvaz/projects/smaqit-adk/installer
make clean build test
```

Expected output: Binary at `dist/smaqit-adk-dev`

### 3. Test Installation

```bash
cd /home/ruifrvaz/projects/smaqit-adk/installer
make test
```

This will:
- Create test-project/
- Run `smaqit-adk init`
- Validate 15 files exist
- Verify no product-specific contamination

### 4. ✅ Initialize Git Repository

```bash
cd /home/ruifrvaz/projects/smaqit-adk
git init
git add .
git commit -m "feat: initial ADK release - generic agent development kit

- Generic framework files (5): SMAQIT, AGENTS, TEMPLATES, ARTIFACTS, PROMPTS
- Generic agent templates (3): base, specification, implementation
- Compiled rules (3): base, specification, implementation - cleaned of product coupling
- Level agents (3): L0 + L1 fully cleaned, L2 partial (body cleaning pending)
- Renamed SDK -> ADK throughout
- Self-contained installer (smaqit-adk CLI)
- ADK-focused documentation and wiki

Version: adk-v0.1.0"
```

### 5. ✅ Create GitHub Repository

```bash
cd /home/ruifrvaz/projects/smaqit-adk
gh repo create ruifrvaz/smaqit-adk --private --source=. --push
```

### 6. ✅ Tag Initial Release

```bash
git tag adk-v0.1.0
git push origin adk-v0.1.0
```

This triggers `.github/workflows/release.yml` to build binaries.

## Key Decisions Made

1. **Repository location:** `/home/ruifrvaz/projects/smaqit-adk` (parallel to smaQit)
2. **Visibility:** Private initially
3. **Version:** Starting at `adk-v0.1.0`
4. **Original repo:** `/home/ruifrvaz/projects/smaqit` left **completely untouched**
5. **Framework scope:** Excluded LAYERS.md and PHASES.md (smaQit product-specific)
6. **No external dependencies:** ADK installer uses only Go stdlib
7. **Installer is self-contained:** No `../` references, embeds from local structure

## Architecture

### ADK Scope (Generic)

**Includes:**
- Core principles (WHY/WHAT)
- Generic agent patterns (base, specification, implementation)
- Generic compilation rules
- Level agents (L0/L1/L2 meta-compilation)
- new-agent prompt template

**Excludes:**
- Layer models (business/functional/stack/infrastructure/coverage)
- Phase models (develop/deploy/validate)
- Spec templates (outputs of compiled agents)
- Product agents (compiled outputs)

### Product Scope (smaQit)

**Remains in `/home/ruifrvaz/projects/smaqit`:**
- 8 compiled agents (business, functional, stack, infrastructure, coverage, development, deployment, validation)
- 5 spec templates
- 8 prompts
- Product installer
- LAYERS.md, PHASES.md framework files
- Layer/phase specific compilation rules

## Known Issues

1. **Level agents need cleaning** - Currently contain smaQit product references
2. **Framework files may have contamination** - Not yet reviewed for product-specific content
3. **Not yet tested** - Build and installation untested
4. **No Git history** - Clean slate, loses original commit history

## Session Context

This handover created during **smaQit Task 075: Dual Release Architecture** implementation. Original task was to split ADK and product installers within smaQit monorepo, but assessment revealed deep entanglement requiring full repository split.

**Original smaQit repo remains on Task 075** (in progress), will complete after ADK is stable.

## Commands for New Session

```bash
# Navigate to ADK
cd /home/ruifrvaz/projects/smaqit-adk

# Review Level agents for cleaning
code agents/smaqit.L0.agent.md agents/smaqit.L1.agent.md agents/smaqit.L2.agent.md

# After cleaning, test build
cd installer && make clean build test

# Initialize repo (after successful test)
cd /home/ruifrvaz/projects/smaqit-adk
git init
git add .
git commit -m "Initial ADK extraction"
gh repo create ruifrvaz/smaqit-adk --private --source=. --push

# Tag release (after build/push success)
git tag adk-v0.1.0
git push origin adk-v0.1.0
```

## Related Files

- Original smaQit repo: `/home/ruifrvaz/projects/smaqit`
- Task tracking: `/home/ruifrvaz/projects/smaqit/docs/tasks/PLANNING.md` (Task 075)
- Session history: Will be documented in `/home/ruifrvaz/projects/smaqit/docs/history/048_*.md`

## Questions to Resolve

1. Should framework files (SMAQIT.md, AGENTS.md, etc.) be reviewed for smaQit product contamination?
2. After ADK is stable, does smaQit product eventually import ADK as Go module dependency?
3. Should ADK version reach 1.0.0 quickly (stable framework) or iterate in 0.x (evolving)?

## Success Criteria

✅ ADK builds successfully
✅ ADK installs 15 files (no product-specific files)
✅ Level agents contain no smaQit product references
✅ GitHub repo created and pushed
✅ Release workflow produces binaries
✅ Integration test passes

Then return to smaQit repo to complete Task 075.

## Updates After Initial Handover

### install.sh Improved (2026-02-05)

Replaced simple install.sh with better version from smaQit/install-sdk.sh:
- Added colored output (info/warn/error helpers)
- Added version handling: `latest`, `prerelease`, or specific `adk-vX.Y.Z`
- Added installation verification
- Better error messages and user guidance
- Updated REPO variable to `ruifrvaz/smaqit-adk`

Usage examples:
```bash
# Install latest stable ADK
curl -fsSL https://raw.githubusercontent.com/ruifrvaz/smaqit-adk/main/install.sh | bash

# Install latest prerelease
curl -fsSL https://raw.githubusercontent.com/ruifrvaz/smaqit-adk/main/install.sh | SMAQIT_ADK_VERSION=prerelease bash

# Install specific version
curl -fsSL https://raw.githubusercontent.com/ruifrvaz/smaqit-adk/main/install.sh | SMAQIT_ADK_VERSION=adk-v0.1.0 bash
```
- Moved extending-smaQit.md wiki from smaQit to ADK (this is ADK-specific content)
