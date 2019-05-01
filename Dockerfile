FROM golang:1.12.3-alpine3.9 AS builder

ENV APP_DIR /amznode
WORKDIR $APP_DIR

RUN apk add --no-cache git=2.20.1-r0

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build ./cmd/amznode

FROM alpine:3.9

COPY --from=builder /amznode/amznode /amznode

CMD ["/amznode"]
