run: build
	@./bin/cish

build:
	@go build -o ./bin/cish

test:
	@go test ./... -v
