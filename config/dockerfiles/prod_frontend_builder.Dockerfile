# Compila el frontend Angular. El resultado queda en /frontend/dist y el servicio
# builder_frontend de prod.yml lo publica como artefacto para nginx.
FROM node:24-alpine AS build

WORKDIR /frontend

COPY frontend/package*.json ./
RUN npm ci

COPY frontend/ ./

RUN npm run build
