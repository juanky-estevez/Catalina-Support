FROM golang:1.25-alpine AS manager-build
WORKDIR /src
COPY ai-manager/go.mod ai-manager/main.go ./
RUN CGO_ENABLED=0 go build -o /ai-manager .

FROM ghcr.io/ggml-org/llama.cpp:server-b11206
COPY --from=manager-build /ai-manager /usr/local/bin/ai-manager
ENTRYPOINT ["/usr/local/bin/ai-manager"]
