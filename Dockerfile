FROM node:22-alpine AS web-build

WORKDIR /app/web

COPY web/package.json web/package-lock.json ./

RUN npm ci

COPY web/ ./

RUN npm run build


FROM golang:1.25-alpine AS go-build

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY embed.go ./
COPY --from=web-build /app/web/dist ./web/dist

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /server ./cmd/server


FROM alpine:3.21

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=go-build /server ./server

ENV PORT=8080

EXPOSE 8080

CMD ["./server"]
