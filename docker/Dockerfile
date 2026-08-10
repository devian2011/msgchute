FROM node:26-alpine AS frontend-builder
WORKDIR /web
COPY ./web/package*.json ./
RUN npm ci
COPY ./web/ .
RUN npm run build

FROM golang:latest AS builder

RUN mkdir -p /app
WORKDIR /app
COPY /app/go.mod /app/go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./app/cmd/main/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates &&  \
    mkdir /app &&  \
    mkdir -p /app/plugins/providers &&  \
    mkdir -p /app/plugins/auth && \
    mkdir -p /app/migrations && \
    mkdir -p /app/dashboard

COPY ./migrations /app/migrations
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=frontend-builder /web/dist /app/dashboard

CMD ["./main"]
