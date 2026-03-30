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
	@rm -rf agents && mkdir -p agents
	@cp ../agents/*.agent.md agents/
	@rm -rf skills && mkdir -p skills/smaqit.new-agent && mkdir -p skills/smaqit.new-skill
	@cp ../skills/smaqit.new-agent/SKILL.md skills/smaqit.new-agent/
	@cp ../skills/smaqit.new-skill/SKILL.md skills/smaqit.new-skill/
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

# Smoke-test installation scaffold (preserved from original 'test' target)
.PHONY: test-scaffold
test-scaffold: build
	@echo "Testing ADK installation..."
	@mkdir -p test-project
	@cd test-project && ../$(DIST_DIR)/$(DEV_BINARY_NAME) init
	@echo "✓ Test installation complete"
	@echo "Validating structure..."
	@test -d test-project/.smaqit/framework || (echo "✗ Missing .smaqit/framework" && exit 1)
	@test -d test-project/.smaqit/templates/agents || (echo "✗ Missing .smaqit/templates/agents" && exit 1)
	@test -d test-project/.github/agents || (echo "✗ Missing .github/agents" && exit 1)
	@test -d test-project/.github/skills/smaqit.new-agent || (echo "✗ Missing .github/skills/smaqit.new-agent" && exit 1)
	@echo "✓ Structure validation passed"
	@rm -rf test-project
	@echo "✓ Test complete"

# Layer 1 + 2: black-box unit tests + structural validation
.PHONY: test
test: build
	@cd $(REPO_ROOT)/tests && SMAQIT_ADK_BIN=$(CURDIR)/$(DIST_DIR)/$(DEV_BINARY_NAME) go test ./unit/... ./structural/...

# Layer 3: behavioral evaluations (requires an explicit token — no VS Code auth fallback)
# Usage: GH_TOKEN=$(gh auth token) make evals
.PHONY: evals
evals: build
	@if [ -z "$$COPILOT_GITHUB_TOKEN" ] && [ -z "$$GH_TOKEN" ] && [ -z "$$GITHUB_TOKEN" ]; then \
		echo "error: no GitHub token found. Run: GH_TOKEN=\$$(gh auth token) make evals"; \
		exit 1; \
	fi
	@cd $(REPO_ROOT)/tests && go run ./evals/runner/... $(REPO_ROOT)/tests/evals/

# All layers
.PHONY: test-all
test-all: test evals

# Help
.PHONY: help
help:
	@echo "smaqit-adk installer build system"
	@echo ""
	@echo "Targets:"
	@echo "  build      Build for current platform (default)"
	@echo "  build-all  Build for all platforms"
	@echo "  test       Test installation"
	@echo "  clean      Remove build artifacts"
	@echo "  version    Show version that would be built"
	@echo "  help       Show this help"

.DEFAULT_GOAL := build

