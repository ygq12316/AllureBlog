# ===== Stage 1: Build Vue SPA =====
FROM node:20-alpine AS node-builder

ARG HTTP_PROXY
ARG HTTPS_PROXY
ENV http_proxy=$HTTP_PROXY https_proxy=$HTTPS_PROXY

RUN npm config set registry https://registry.npmmirror.com

WORKDIR /build/web

COPY web/package.json ./
RUN npm install

COPY web/ ./
RUN npm run build

# ===== Stage 2: Compile Go backend =====
FROM golang:alpine AS go-builder

ARG HTTP_PROXY
ARG HTTPS_PROXY
ENV http_proxy=$HTTP_PROXY https_proxy=$HTTPS_PROXY

ENV GOPROXY=https://goproxy.cn,direct

RUN apk add --no-cache gcc musl-dev

WORKDIR /build/server

COPY server/go.mod server/go.sum ./
RUN go mod download

COPY server/ ./

RUN CGO_ENABLED=1 GOOS=linux go build -o /blog ./cmd/server

# ===== Stage 3: Runtime =====
FROM alpine:3.21

ARG HTTP_PROXY
ARG HTTPS_PROXY
ENV http_proxy=$HTTP_PROXY https_proxy=$HTTPS_PROXY

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=go-builder /blog /app/blog
COPY --from=node-builder /build/web/dist /app/web/dist

RUN mkdir -p /app/web/static/uploads /app/data

EXPOSE 8080

ENV DB_PATH=/app/data/blog.db PORT=8080 GIN_MODE=release

CMD ["/app/blog"]
