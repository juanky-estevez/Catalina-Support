# Compila el backend a un binario estático. No se ejecuta: deja el binario en /build
# y el servicio builder_backend de prod.yml lo publica como artefacto.
FROM golang:1.27-alpine AS build

RUN apk add --no-cache git

WORKDIR /src

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags "-s -w" -o /build/catalina-support .
