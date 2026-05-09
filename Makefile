.PHONY: build run test format clean

build:
	mkdir -p ./bin && go build -o ./bin/nexo ./cmd/nexo

run: build
	./bin/nexo

test:
	go test ./...

format:
	go fmt ./...

clean:
	rm -rf ./bin