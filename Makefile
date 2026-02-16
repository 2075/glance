PROJECT    = Glance.xcodeproj
SCHEME     = Glance
DEST       = platform=macOS
BUILD_DIR  = $(shell xcodebuild -scheme $(SCHEME) -showBuildSettings 2>/dev/null | grep '^\s*BUILT_PRODUCTS_DIR' | awk '{print $$NF}')
APP_NAME   = Glance.app
PLUGIN_ID  = io.2075.Glance.QLPlugin
INSTALL_DIR = $(HOME)/Applications

# ── Build ────────────────────────────────────────────────────────────

.PHONY: build
build: ## Build debug configuration
	xcodebuild build \
		-project $(PROJECT) \
		-scheme $(SCHEME) \
		-destination '$(DEST)' \
		-quiet

.PHONY: release
release: ## Build release configuration
	xcodebuild build \
		-project $(PROJECT) \
		-scheme $(SCHEME) \
		-destination '$(DEST)' \
		-configuration Release \
		-quiet

# ── Test ─────────────────────────────────────────────────────────────

.PHONY: test
test: ## Run unit tests
	xcodebuild test \
		-project $(PROJECT) \
		-scheme $(SCHEME) \
		-destination '$(DEST)' \
		-quiet

# ── Install / Uninstall ─────────────────────────────────────────────

.PHONY: install
install: build ql-reset ## Build, copy to ~/Applications, and register the QL extension
	@mkdir -p "$(INSTALL_DIR)"
	@rm -rf "$(INSTALL_DIR)/$(APP_NAME)"
	@cp -R "$(BUILD_DIR)/$(APP_NAME)" "$(INSTALL_DIR)/$(APP_NAME)"
	@echo "Installed to $(INSTALL_DIR)/$(APP_NAME)"
	@open "$(INSTALL_DIR)/$(APP_NAME)"
	@sleep 2
	@pluginkit -e use -i $(PLUGIN_ID) 2>/dev/null || true
	@echo "QL extension registered — open a supported file in Finder and press Space"

.PHONY: uninstall
uninstall: ## Remove from ~/Applications and reset QL cache
	@rm -rf "$(INSTALL_DIR)/$(APP_NAME)"
	@pluginkit -e ignore -i $(PLUGIN_ID) 2>/dev/null || true
	@qlmanage -r >/dev/null 2>&1 || true
	@echo "Uninstalled $(APP_NAME)"

# ── Run ──────────────────────────────────────────────────────────────

.PHONY: run
run: build ## Build and launch the app
	@open "$(BUILD_DIR)/$(APP_NAME)"

# ── Quick Look helpers ───────────────────────────────────────────────

.PHONY: ql-reset
ql-reset: ## Reset Quick Look daemon and caches
	@qlmanage -r >/dev/null 2>&1
	@qlmanage -r cache >/dev/null 2>&1
	@echo "Quick Look caches reset"

.PHONY: ql-preview
ql-preview: ## Preview README.md via qlmanage (pass FILE= to override)
	qlmanage -p $(or $(FILE),README.md)

.PHONY: ql-status
ql-status: ## Show registered Glance QL extension info
	@pluginkit -mAvvv -p com.apple.quicklook.preview 2>&1 | grep -A10 "$(PLUGIN_ID)" || echo "QL extension not registered"

# ── Release ──────────────────────────────────────────────────────────

VERSION = $(shell sed -n 's/.*MARKETING_VERSION = \(.*\);/\1/p' $(PROJECT)/project.pbxproj | head -1 | tr -d ' ')

.PHONY: tag
tag: ## Create and push a git tag for the current MARKETING_VERSION (triggers GitHub release)
	@echo "Tagging v$(VERSION)..."
	git tag -a "v$(VERSION)" -m "Release $(VERSION)"
	git push origin "v$(VERSION)"
	@echo "Pushed tag v$(VERSION) — GitHub Actions will build and create the release"

.PHONY: retag
retag: ## Delete and re-push the current version tag to retrigger the release
	git tag -fa "v$(VERSION)" -m "Release $(VERSION)"
	git push origin "v$(VERSION)" --force
	@echo "Re-tagged v$(VERSION) — GitHub Actions will rebuild the release"

# ── Clean ────────────────────────────────────────────────────────────

.PHONY: clean
clean: ## Remove build artifacts
	xcodebuild clean \
		-project $(PROJECT) \
		-scheme $(SCHEME) \
		-quiet
	@echo "Build artifacts cleaned"

# ── Help ─────────────────────────────────────────────────────────────

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
