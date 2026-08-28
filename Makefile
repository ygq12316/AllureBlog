.PHONY: run build test vet lint clean tidy

# web 根由服务自探（仓库根 ./web，server/ 内回退 ../web），make 目标统一在根执行
run:
	cd server && go run ./cmd/server

build:
	go build -o blog.exe ./server/cmd/server

test:
	cd server && go test ./...

vet:
	cd server && go vet ./...

lint:
	cd server && golangci-lint run

clean:
	rm -f blog.exe

tidy:
	cd server && go mod tidy
