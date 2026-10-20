.PHONY: all build run server wasm test clean

APP_NAME = mimarea

all: build wasm

build:
	go build -o $(APP_NAME) main.go

wasm:
	mkdir -p dist/css dist/js dist/icons
	GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o dist/tides.wasm cmd/wasm/main.go
	cp /opt/homebrew/Cellar/go/1.26.5/libexec/lib/wasm/wasm_exec.js dist/js/wasm_exec.js
	cp web/css/style.css dist/css/style.css
	go run scratch/gen_icon.go
	@echo "✅ Carpeta dist/ compilada y lista para Cloudflare / cdmon."

run: build
	./$(APP_NAME)

server: build
	./$(APP_NAME) -headless

test:
	go test ./... -v

clean:
	rm -f $(APP_NAME)
