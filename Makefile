APP_NAME := wso-stateless-grpc
DOCKER_IMAGE := bartwild/$(APP_NAME):latest

.PHONY: build run docker-build docker-run clean, generate-proto

build:
	go build -o bin/server ./cmd/main

run: build
	./bin/server

docker-build:
	docker build -t $(DOCKER_IMAGE) .

docker-run:
	docker run --rm -p 50051:50051 -p 8081:8081 $(DOCKER_IMAGE)

clean:
	rm -rf bin/

generate-proto:
	protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/dot_product/dot_product.proto