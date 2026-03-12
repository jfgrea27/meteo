set dotenv-load

# List available recipes
default:
    @just --list

# Run all Go tests
test:
    cd src && go test ./...

# Run all Go tests
test-coverage:
    cd src && go test ./... -cover

# Build a specific service (api, consumer, producer)
build SVC:
    cd src && go build -o ../bin/{{SVC}} ./cmd/{{SVC}}

docker-build SVC:
    docker build -f Dockerfile --build-arg SVC={{SVC}} -t {{SVC}}:latest src

docker-run SVC: 
    just docker-build {{SVC}}
    docker run -it --rm {{SVC}}:latest

docker-up:
    cd src && docker compose down -v
    cd src && docker compose up --build

# Build all services
build-all:
    just build api
    just build consumer
    just build producer

# Run
run SVC:
    cd src && go run ./cmd/{{SVC}}

proto:
    cd src && protoc \
        --go_out=. \
        --go_opt=paths=source_relative \
        --go-grpc_out=. \
        --go-grpc_opt=paths=source_relative \
        proto/*.proto