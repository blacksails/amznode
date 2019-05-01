BINARY=amznode

all: deps test build
deps:
	go mod download
test: 
	docker-compose up -d
	go test ./...
build: 
	go build ./cmd/$(BINARY)

clean: clean-go clean-binary clean-compose
clean-go:
	go clean
clean-binary:
	rm -f $(BINARY)
clean-compose:
	docker-compose down

run:
	go run ./cmd/$(BINARY)/main.go

docker:
	docker build -t blacksails/$(BINARY) .

compose: compose-stop compose-build compose-up
compose-stop:
	docker-compose stop
compose-build:
	docker-compose build
compose-up:
	docker-compose up
