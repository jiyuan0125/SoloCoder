.PHONY: all build-backend build-frontend run-backend run-frontend clean

all: build-backend build-frontend

build-backend:
	cd backend && go build -o ../bin/server ./cmd/server

build-cli:
	cd backend && go build -o ../bin/mqs-cli ./cmd/cli

build-frontend:
	cd frontend && npm install && npm run build

run-backend:
	cd backend && go run ./cmd/server

run-frontend:
	cd frontend && npm run dev

clean:
	rm -rf bin
