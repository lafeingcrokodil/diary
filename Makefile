.PHONY: run test vet

run:
	go run ./cmd/diary

test:
	go test ./...

vet:
	go vet ./...
