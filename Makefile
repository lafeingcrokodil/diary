.PHONY: run test vet lint

run:
	go run ./cmd/diary

test:
	go test -race ./...

vet:
	go vet ./...

lint:
	golangci-lint run ./...
