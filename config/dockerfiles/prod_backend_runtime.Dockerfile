# Runtime del backend: sin Go, sin air y sin código fuente. El binario entra por
# volumen desde /root/prod/catalina-support (ver prod.yml).
FROM alpine:3.21

# ca-certificates y tzdata los necesitan las conexiones TLS y las fechas locales.
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
