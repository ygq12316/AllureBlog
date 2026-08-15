# ===== Stage 1: Build Vue SPA =====
FROM node:20-alpine AS node-builder

ARG HTTP_PROXY
ARG HTTPS_PROXY
ENV http_proxy=$HTTP_PROXY https_proxy=$HTTPS_PROXY

RUN npm config set registry https://registry.npmmirror.com

WORKDIR /app/web/admin

COPY web/admin/package.json ./
RUN npm install

COPY web/admin/ ./
RUN npm run build

# ===== Stage 2: Compile Go backend =====
FROM golang:alpine AS go-builder

ARG HTTP_PROXY
ARG HTTPS_PROXY
ENV http_proxy=$HTTP_PROXY https_proxy=$HTTPS_PROXY

ENV GOPROXY=https://goproxy.cn,direct

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Overlay Vue dist from stage 1 into Go source tree
COPY --from=node-builder /app/web/admin/dist web/admin/dist

RUN CGO_ENABLED=1 GOOS=linux go build -o /blog ./cmd/server

# ===== Stage 3: Runtime =====
FROM alpine:3.21

ARG HTTP_PROXY
ARG HTTPS_PROXY
ENV http_proxy=$HTTP_PROXY https_proxy=$HTTPS_PROXY

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=go-builder /blog /app/blog
COPY --from=node-builder /app/web/admin/dist /app/web/admin/dist

RUN mkdir -p /app/web/static/uploads /app/data

EXPOSE 8080

ENV DB_PATH=/app/data/blog.db PORT=8080 GIN_MODE=release

CMD ["/app/blog"]
