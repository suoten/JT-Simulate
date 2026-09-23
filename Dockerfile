FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o jt-simulate ./cmd/jt-simulate

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Shanghai
WORKDIR /app
COPY --from=builder /app/jt-simulate .
COPY configs/ ./configs/
COPY data/ ./data/
EXPOSE 8095
CMD ["./jt-simulate", "serve"]
