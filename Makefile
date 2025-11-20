NIX_PORTABLE_VERSION = v012
NIX_PORTABLE_URL = https://github.com/DavHau/nix-portable/releases/download/$(NIX_PORTABLE_VERSION)/nix-portable-x86_64
NIX_PORTABLE_PATH = src/nix/nix-portable-binary

download-nix-portable:
	@if [ ! -f $(NIX_PORTABLE_PATH) ]; then \
		echo "Downloading nix-portable..."; \
		curl -L $(NIX_PORTABLE_URL) -o $(NIX_PORTABLE_PATH); \
		chmod +x $(NIX_PORTABLE_PATH); \
	fi

build: download-nix-portable
	CGO_ENABLED=0 go build -o bin/scriptkiller ./src

run: build
	./bin/scriptkiller $(ARGS) .

test:
	CGO_ENABLED=0 go test

