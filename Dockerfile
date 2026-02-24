# syntax=docker/dockerfile:1

FROM node:20-alpine AS web-build
WORKDIR /app
RUN apk add --no-cache git
COPY web/package.json web/pnpm-lock.yaml web/package-lock.json* web/ ./web/
RUN cd web && \
    if [ -f pnpm-lock.yaml ]; then npm i -g pnpm && pnpm install; \
    else npm install; fi
COPY web/ ./web/
RUN cd web && npm run build

FROM golang:1.24.2-alpine AS go-build
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
COPY --from=web-build /app/web/out /app/static/out
RUN go build -o nodehub ./cmd/nodehub

FROM alpine:3.20
WORKDIR /app
COPY --from=go-build /app/nodehub ./nodehub
VOLUME ["/data"]
ENV NODEHUB_SERVER_HOST=0.0.0.0
ENV NODEHUB_SERVER_PORT=8080
EXPOSE 8080
CMD ["./nodehub", "-c", "/data/config.json"]
