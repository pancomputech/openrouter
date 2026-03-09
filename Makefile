.PHONY: test tidy fmt vet build

EXE=openrouter

test: tidy fmt vet
	go test ./...

tidy:
	go mod tidy

fmt:
	go fmt ./...

vet:
	go vet ./...

build:
	go build -o $(EXE) ./cmd/openrouter

clean:
	rm -f $(EXE)
