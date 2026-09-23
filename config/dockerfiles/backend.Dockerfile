# Backend de desarrollo: sólo la herramienta (Go 1.27 + air). El código entra por
# volumen desde la máquina (ver dev.yml), así que esta imagen no lo copia.
FROM golang:1.27-alpine

# git lo necesita "go mod" para resolver dependencias.
RUN apk add --no-cache git

RUN go install github.com/air-verse/air@v1.67.4

WORKDIR /app

EXPOSE 11002

CMD ["air", "-c", ".air.toml"]
