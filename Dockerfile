FROM golang:1.23 AS builder

WORKDIR /app

COPY backend/go.mod backend/go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod/ \
    go mod download

COPY backend/*.go ./
COPY backend/cmd ./cmd
COPY backend/internal ./internal
COPY backend/migrations ./migrations
ENV CGO_ENABLED=1 GOOS=linux GOARCH=arm64
ENV GOCACHE=/root/.cache/go-build
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=cache,target="/root/.cache/go-build" \
    go build -ldflags "-extldflags -static" -o homecloud ./cmd/homecloud/main.go

FROM node:22-slim AS frontend-base
ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN npm i -g pnpm
COPY frontend /app
WORKDIR /app

FROM frontend-base AS frontend-builder
RUN --mount=type=cache,id=pnpm,target=/pnpm/store pnpm install --frozen-lockfile
RUN pnpm run build

FROM debian:bookworm-slim
COPY --from=builder /app/homecloud /app/homecloud
COPY --from=frontend-builder /app/build /app/spa
COPY backend/assets /app/assets
WORKDIR /app
ENTRYPOINT ["./homecloud"]
EXPOSE 1323