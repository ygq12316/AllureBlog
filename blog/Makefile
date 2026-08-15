.PHONY: run build test vet lint clean tidy

run:
	go run ./cmd/server

build:
	go build -o blog.exe ./cmd/server

# 认证配置为必填项（handler 包启动即校验），测试时注入占位值
test:
	JWT_SECRET=test-secret ADMIN_PASSWORD=test-pass go test ./...

vet:
	JWT_SECRET=test-secret ADMIN_PASSWORD=test-pass go vet ./...

lint:
	golangci-lint run

clean:
	rm -f blog.exe

tidy:
	go mod tidy
