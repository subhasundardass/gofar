# ══════════════════════════════════════════════════════════════════════════════
#  GoFar — Project Makefile
#  Run `make` or `make help` to see all available commands.
# ══════════════════════════════════════════════════════════════════════════════

CLI := go run ./cmd/gofar
CSS_IN  = static/assets/css/input.css
CSS_OUT = static/assets/css/app.css

.PHONY: help run build tidy \
        module remove handler service \
        list-modules sync-modules

# ── Default ────────────────────────────────────────────────────────────────────

help:
	@printf "\n"
	@printf "  \033[1mGoFar\033[0m — available commands\n"
	@printf "\n"
	@printf "  \033[33mDevelopment\033[0m\n"
	@printf "    make run                               Start the dev server\n"
	@printf "    make dev                               Start the air\n"
	@printf "    make build                             Compile → bin/server\n"
	@printf "    make tidy                              Tidy go.mod & go.sum\n"
	@printf "    make templ                             Generate templ to go\n"
	@printf "    make css                               Generate css\n"
	@printf "    make css-watch                         Generate css with watch\n"
	
	@printf "\n"
	@printf "  \033[33mScaffold\033[0m\n"
	@printf "    make module  name=<Name>               Create a new module\n"
	@printf "    make remove  name=<Name>               Delete a module\n"
	@printf "    make handler name=<Name> module=<Mod>  Add a handler to a module\n"
	@printf "    make service name=<Name> module=<Mod>  Add a service to a module\n"
	@printf "\n"
	@printf "  \033[33mModules\033[0m\n"
	@printf "    make list-modules                      Show registered modules\n"
	@printf "    make sync-modules                      Fix stale imports in all.go\n"
	@printf "\n"
	@printf "  \033[2mExamples\033[0m\n"
	@printf "    make module  name=Billing\n"
	@printf "    make module  name=BillingNotification\n"
	@printf "    make remove  name=Billing\n"
	@printf "    make handler name=Refund module=Billing\n"
	@printf "    make service name=TaxCalculator module=Billing\n"
	@printf "\n"

# ── Development ────────────────────────────────────────────────────────────────

run:
	go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

tidy:
	go mod tidy

# ── Scaffold ───────────────────────────────────────────────────────────────────
#
#  `make module name=Billing` generates:
#
#    modules/billing/
#    ├── manifest.json
#    ├── module.go
#    ├── routes.go
#    ├── events.go
#    ├── handler/handler.go
#    ├── service/service.go
#    └── repository/repository.go
#
#  And auto-registers in modules/all/all.go.

module:
ifndef name
	@printf "\n  \033[31merror:\033[0m name is required\n"
	@printf "  usage:  make module name=<Name>\n"
	@printf "  example: make module name=Billing\n\n"
	@exit 1
endif
	@$(CLI) make:module name=$(name)

remove:
ifndef name
	@printf "\n  \033[31merror:\033[0m name is required\n"
	@printf "  usage:  make remove name=<Name>\n"
	@printf "  example: make remove name=Billing\n\n"
	@exit 1
endif
ifdef force
	@$(CLI) remove:module name=$(name) force=1
else
	@$(CLI) remove:module name=$(name)
endif

handler:
ifndef name
	@printf "\n  \033[31merror:\033[0m name and module are required\n"
	@printf "  usage:  make handler name=<Name> module=<Module>\n"
	@printf "  example: make handler name=Refund module=Billing\n\n"
	@exit 1
endif
ifndef module
	@printf "\n  \033[31merror:\033[0m module is required\n"
	@printf "  usage:  make handler name=<Name> module=<Module>\n"
	@printf "  example: make handler name=Refund module=Billing\n\n"
	@exit 1
endif
	@$(CLI) make:handler name=$(name) module=$(module)

service:
ifndef name
	@printf "\n  \033[31merror:\033[0m name and module are required\n"
	@printf "  usage:  make service name=<Name> module=<Module>\n"
	@printf "  example: make service name=TaxCalculator module=Billing\n\n"
	@exit 1
endif
ifndef module
	@printf "\n  \033[31merror:\033[0m module is required\n"
	@printf "  usage:  make service name=<Name> module=<Module>\n"
	@printf "  example: make service name=TaxCalculator module=Billing\n\n"
	@exit 1
endif
	@$(CLI) make:service name=$(name) module=$(module)

# ── Modules ────────────────────────────────────────────────────────────────────

list-modules:
	@$(CLI) list:modules

# Run after manually deleting a module folder with rm -rf
# Removes any import in modules/all/all.go whose directory no longer exists
sync-modules:
	@$(CLI) sync:modules

# ── Generate templ components ─────────────────────────────────────────────────
templ:
	templ generate
	@echo "✅ templ generated"

# ── Tailwind CSS ──────────────────────────────────────────────────────────────
css:
	npm run tailwind -- -i $(CSS_IN) -o $(CSS_OUT) --minify
	@echo "✅ css built"

css-watch:
	npm run tailwind -- -i $(CSS_IN) -o $(CSS_OUT) --watch

# -----------------------------------------------------------------------------
# Format code
# -----------------------------------------------------------------------------
fmt:
	go fmt ./...

# -----------------------------------------------------------------------------
# Live reload (requires https://github.com/air-verse/air)
# -----------------------------------------------------------------------------
dev: templ css
	air

# -----------------------------------------------------------------------------
# Clean build artifacts
# -----------------------------------------------------------------------------
clean:
	rm -rf $(BIN_DIR)

# =========================================================
# Kill Port
# =========================================================
kill:
	@echo "Killing process on port 3000..."
	@kill -9 $$(lsof -t -i:3000) 2>/dev/null || true
	@echo "Done"	