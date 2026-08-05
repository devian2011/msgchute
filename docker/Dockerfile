FROM golang:latest AS builder

RUN mkdir -p /app
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates && mkdir /app && mkdir /app/templates
WORKDIR /app
COPY --from=builder /app/main .

CMD ["./main"]
