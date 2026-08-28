.PHONY: run build test vet lint clean tidy

run:
	go run ./cmd/server

build:
	go build -o blog.exe ./cmd/server

# 认证配置注入发生在 main 启动时（requireEnv），handler 包已无包级副作用
test:
	go test ./...

vet:
	go vet ./...

lint:
	golangci-lint run

clean:
	rm -f blog.exe

tidy:
	go mod tidy
