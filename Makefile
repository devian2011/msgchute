include .env

.PHONY: *

# Build project
build: build-server build-smtp build-tg
# Build main server
build-server:
	go build -a -installsuffix cgo -o app ./app/cmd/main/main.go
# Build smtp transport
build-smtp:
	go build -a -installsuffix cgo -o ./plugins/dist/providers/smtp ./plugins/src/providers/smtp/provider.go
# Build telegram transport
build-tg:
	go build -a -installsuffix cgo -o ./plugins/dist/providers/tg ./plugins/src/providers/telegram/provider.go

build-auth:
	go build -a -installsuffix cgo -o ./plugins/dist/auth/base ./plugins/src/auth/base/auth.go

# Test project
test: test-server test-smtp test-tg
# Test server
test-server:
	go test -race ./app/...
# Test smtp provider
test-smtp:
	go test -race ./plugins/src/providers/smtp/...
# Test telegram provider
test-tg:
	go test -race ./plugins/src/providers/telegram/...


lint:
	golangci-lint run ./app/...

migrate:
	docker run --rm -it --network=notifier -v "$(PWD)/migrations:/db/migrations" -e DATABASE_URL=$(APP_DB_DSN) --env-file=.env ghcr.io/amacneil/dbmate $(command)

swagger:
	swag init -d ./app -g ./cmd/main/main.go -o ./app/docs --parseDependency --parseInternal && swag fmt ./app/...

