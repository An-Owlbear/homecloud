FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod/ \
    go mod download

COPY *.go ./
COPY cmd ./cmd
COPY internal ./internal
COPY migrations ./migrations
ENV CGO_ENABLED=1 GOOS=linux GOARCH=arm64
ENV GOCACHE=/root/.cache/go-build
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=cache,target="/root/.cache/go-build" \
    go build -ldflags "-extldflags -static" -o homecloud ./cmd/homecloud/main.go

FROM gcr.io/distroless/static-debian12
COPY --from=builder /app/homecloud /app/homecloud
COPY assets /app
WORKDIR /app
ENTRYPOINT ["./homecloud"]
EXPOSE 1323