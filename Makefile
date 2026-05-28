.PHONY: frontend-install dev build test clean

frontend-install:
	cd frontend && npm install

dev:
	wails dev

build:
	wails build

test:
	go test ./...

clean:
	rm -rf build/bin frontend/dist
