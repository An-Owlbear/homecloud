FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
ENV CGO_ENABLED=1 GOOS=linux GOARCH=arm64
RUN go build -a -ldflags "-extldflags -static" -o homecloud ./cmd/homecloud/main.go

FROM gcr.io/distroless/static-debian12
COPY --from=builder /app/homecloud /app/homecloud
WORKDIR /app
ENTRYPOINT ["./homecloud"]
EXPOSE 1323