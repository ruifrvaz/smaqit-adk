# smaqit-adk installer build system

# Detect version from git tags, fallback to commit hash if no tags exist
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# Build configuration
BINARY_NAME = smaqit-adk
DEV_BINARY_NAME = smaqit-adk-dev
DIST_DIR = dist
LDFLAGS = -ldflags "-X main.Version=$(VERSION)"

# Copy embedded files (required for go:embed at compile time)
.PHONY: prepare
prepare:
	@echo "Copying ADK embedded files..."
	@cp -r ../framework .
	@rm -rf templates/agents && mkdir -p templates/agents/compiled
	@cp ../templates/agents/*.template.md templates/agents/
	@cp ../templates/agents/compiled/*.rules.md templates/agents/compiled/
	@GOOS= GOARCH= CGO_ENABLED= go run generate-agents.go
	@rm -rf skills && mkdir -p skills/smaqit.new-principle
	@cp ../skills/smaqit.new-principle/SKILL.md skills/smaqit.new-principle/
	@mkdir -p skills/smaqit.create-agent && mkdir -p skills/smaqit.create-skill/scripts
	@cp ../skills/smaqit.create-agent/SKILL.md skills/smaqit.create-agent/
	@cp ../skills/smaqit.create-skill/SKILL.md skills/smaqit.create-skill/
	@cp ../skills/smaqit.create-skill/scripts/validate-skill.go skills/smaqit.create-skill/scripts/
	@mkdir -p skills/smaqit.bench-run && mkdir -p skills/smaqit.bench-scaffold
	@cp ../skills/smaqit.bench-run/SKILL.md skills/smaqit.bench-run/
	@cp ../skills/smaqit.bench-scaffold/SKILL.md skills/smaqit.bench-scaffold/
	@rm -rf templates/skills && mkdir -p templates/skills/compiled
	@cp ../templates/skills/base-skill.template.md templates/skills/
	@cp ../templates/skills/compiled/skill.rules.md templates/skills/compiled/

# Default target: build for current platform
.PHONY: build
build: prepare
	@echo "Building $(DEV_BINARY_NAME) version $(VERSION) for current platform..."
	@mkdir -p $(DIST_DIR)
	@go build $(LDFLAGS) -o $(DIST_DIR)/$(DEV_BINARY_NAME) .
	@echo "Built: $(DIST_DIR)/$(DEV_BINARY_NAME)"
	@# Pre-built for the .smaqit/bench dogfood suite's validator-passes command
	@# grader: `go run` resolves through a Snap-packaged toolchain
	@# (/snap/bin/go -> /usr/bin/snap) on some dev machines, and Bench's
	@# process-group isolation for reliable timeout/kill handling is
	@# incompatible with Snap's own confinement locking there ("race condition
	@# detected, snap-run can only retry once"). A plain compiled binary sidesteps
	@# Snap entirely at grading time. See .smaqit/bench/README.md.
	@go build -o $(DIST_DIR)/validate-skill ../skills/smaqit.create-skill/scripts/validate-skill.go

# Build for all platforms
.PHONY: build-all
build-all: prepare
	@echo "Building ADK for all platforms..."
	@mkdir -p $(DIST_DIR)
	@echo "Building for linux/amd64..."
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)_linux_amd64 .
	@echo "Building for darwin/amd64..."
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)_darwin_amd64 .
	@echo "Building for darwin/arm64..."
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)_darwin_arm64 .
	@echo "Building for windows/amd64..."
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)_windows_amd64.exe .
	@echo "✓ Built all platforms"

# Show version that would be built
.PHONY: version
version:
	@echo $(VERSION)

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(DIST_DIR)
	@rm -rf framework templates agents skills
	@echo "✓ Clean complete"

REPO_ROOT = ..

# Smoke-test global installation scaffold (preserved from original 'test'
# target). HOME is sandboxed to test-project/ so this never touches the
# real developer machine's global agent/skill directories.
.PHONY: test-scaffold
test-scaffold: build
	@echo "Testing ADK global installation..."
	@mkdir -p test-project
	@HOME=$(CURDIR)/test-project $(CURDIR)/$(DIST_DIR)/$(DEV_BINARY_NAME) --install-global
	@echo "✓ Test installation complete"
	@echo "Validating structure..."
	@test -f test-project/.claude/agents/smaqit-L2.md || (echo "✗ Missing .claude/agents/smaqit-L2.md" && exit 1)
	@test -f test-project/.codex/agents/smaqit.L2.toml || (echo "✗ Missing .codex/agents/smaqit.L2.toml" && exit 1)
	@test -d test-project/.agents/skills/smaqit.create-agent || (echo "✗ Missing .agents/skills/smaqit.create-agent" && exit 1)
	@test -d test-project/.claude/skills/smaqit.create-agent || (echo "✗ Missing .claude/skills/smaqit.create-agent" && exit 1)
	@test ! -d test-project/.copilot/agents || (echo "✗ Copilot agent directory should not be created" && exit 1)
	@echo "✓ Structure validation passed"
	@rm -rf test-project
	@echo "✓ Test complete"

# Layer 1 + 2: black-box unit tests + structural validation
.PHONY: test
test: build
	@cd $(REPO_ROOT)/src && go test ./...
	@go test ./...
	@cd $(REPO_ROOT)/tests && SMAQIT_ADK_BIN=$(CURDIR)/$(DIST_DIR)/$(DEV_BINARY_NAME) go test ./unit/... ./structural/...

.PHONY: test-bench-examples
test-bench-examples: build
	@$(CURDIR)/$(DIST_DIR)/$(DEV_BINARY_NAME) bench validate $(REPO_ROOT)/examples/bench/single-eval/bench.yaml
	@$(CURDIR)/$(DIST_DIR)/$(DEV_BINARY_NAME) bench validate $(REPO_ROOT)/examples/bench/multi-variant/bench.yaml
	@$(CURDIR)/$(DIST_DIR)/$(DEV_BINARY_NAME) bench validate $(REPO_ROOT)/examples/bench/process-harness/bench.yaml
	@$(CURDIR)/$(DIST_DIR)/$(DEV_BINARY_NAME) bench suite validate $(REPO_ROOT)/.smaqit/bench

# Layer 3: HarnessBench dogfood suite, run against the authenticated local
# Codex CLI (`codex exec`). Requires `codex` on PATH and already logged in —
# there is no token/env-var handoff here, unlike the legacy Copilot runner
# this replaced; Codex manages its own local auth. See .smaqit/bench/README.md.
.PHONY: evals
evals: build
	@$(CURDIR)/$(DIST_DIR)/$(DEV_BINARY_NAME) bench suite run $(REPO_ROOT)/.smaqit/bench

# All layers. Deliberately excludes `evals`: it drives a live, credentialed
# Codex CLI and must stay an explicit, separate step — see .smaqit/bench/README.md.
.PHONY: test-all
test-all: test

# Help
.PHONY: help
help:
	@echo "smaqit-adk installer build system"
	@echo ""
	@echo "Targets:"
	@echo "  build      Build for current platform (default)"
	@echo "  build-all  Build for all platforms"
	@echo "  test       Test source module, installer wrapper, and installation"
	@echo "  test-bench-examples  Validate benchmark examples and the .smaqit/bench dogfood suite"
	@echo "  evals      Run the HarnessBench dogfood suite live against the authenticated Codex CLI"
	@echo "  test-all   Run test (evals stays separate — see 'evals' above)"
	@echo "  clean      Remove build artifacts"
	@echo "  version    Show version that would be built"
	@echo "  help       Show this help"

.DEFAULT_GOAL := build
