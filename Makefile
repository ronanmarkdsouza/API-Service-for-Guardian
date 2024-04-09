build:
	@go build -o ./bin/api-service-for-guardian_linux ./cmd/main/main.go

run:
	@go run ./cmd/main/main.go

test:
	@go test ./...