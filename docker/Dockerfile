FROM golang:latest as builder

RUN mkdir -p /app
WORKDIR /app
COPY /app/go.mod /app
RUN go mod download
COPY /app /app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./app/cmd/main/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates && mkdir /app && mkdir /app/templates
WORKDIR /app
COPY --from=builder /app/main .

CMD ["./main"]
