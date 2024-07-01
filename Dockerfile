FROM golang:1.22-alpine AS builder

WORKDIR /go/src/sso

RUN apk add upx
RUN apk add --no-cache git bash make gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -ldflags="-s -w" -o /go/bin/sso ./cmd/sso/main.go
RUN upx -9 go/bin/sso

FROM alpine:latest AS runner

COPY --from=builder /go/bin/sso ./
COPY config/local.yaml /config/local.yaml

ENV CONFIG_PATH=/config/local.yaml

EXPOSE 4404

CMD ["./sso"]