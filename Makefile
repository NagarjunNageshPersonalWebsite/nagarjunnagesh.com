.PHONY: setup build build-css dev clean

NODE ?= node
NPM ?= npm
TAILWIND := npx tailwindcss

setup:
	$(NPM) ci

build-css:
	$(TAILWIND) -i ./tailwind-input.css -o ./frontend/assets/css/tailwind.css --minify

build: setup build-css

dev:
	$(TAILWIND) -i ./tailwind-input.css -o ./frontend/assets/css/tailwind.css --watch

clean:
	rm -rf node_modules
	rm -f frontend/assets/css/tailwind.css
