# vibeklp build helpers
#
# Tailwind is generated from web/static/css/tailwind-input.css into
# web/static/css/tailwind.css using the standalone Tailwind CLI. The CLI
# binary lives under build/tools/ (gitignored) and is downloaded on demand
# so contributors don't need Node installed.

TAILWIND_VERSION := v3.4.17
TAILWIND_BIN     := build/tools/tailwindcss
TAILWIND_URL     := https://github.com/tailwindlabs/tailwindcss/releases/download/$(TAILWIND_VERSION)/tailwindcss-linux-x64

CSS_INPUT  := web/static/css/tailwind-input.css
CSS_OUTPUT := web/static/css/tailwind.css

.PHONY: css css-watch server crawler clean-css

css: $(CSS_OUTPUT)

$(TAILWIND_BIN):
	@mkdir -p $(dir $@)
	@echo "↳ downloading tailwindcss $(TAILWIND_VERSION)"
	@curl -sSL -o $@ $(TAILWIND_URL)
	@chmod +x $@

$(CSS_OUTPUT): $(CSS_INPUT) tailwind.config.js $(TAILWIND_BIN) $(shell find web/templates web/static/js -type f 2>/dev/null)
	@$(TAILWIND_BIN) -i $(CSS_INPUT) -o $(CSS_OUTPUT) --minify

css-watch: $(TAILWIND_BIN)
	@$(TAILWIND_BIN) -i $(CSS_INPUT) -o $(CSS_OUTPUT) --watch

server: css
	go run ./cmd/server

crawler:
	go run ./cmd/crawler

clean-css:
	rm -f $(CSS_OUTPUT)
