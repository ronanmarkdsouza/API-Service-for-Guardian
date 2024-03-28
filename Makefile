build:
	@go build -o ./bin/app.exe ./cmd/main/main.go

run:
	@go run ./cmd/main/main.go

test:
	@go test ./...