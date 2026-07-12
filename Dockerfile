# --- build stage ---
FROM golang:1.26-bookworm AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -ldflags="-s -w" -o /out/crdledger ./cmd

# --- runtime stage ---
FROM debian:bookworm-slim
WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/* \
	&& useradd -m -u 1000 appuser

COPY --from=build /out/crdledger ./crdledger
COPY templates ./templates
COPY static ./static

RUN mkdir -p /app/static/uploads && chown -R appuser:appuser /app
USER appuser

ENV PORT=8080
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s CMD wget -qO- http://localhost:8080/healthz || exit 1

ENTRYPOINT ["./crdledger"]
