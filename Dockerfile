FROM golang:alpine AS builder

RUN apk update && apk add --no-cache \
  musl-dev \
  gcc \
  build-base

ENV GO111MODULE=on \
  CGO_ENABLED=1 \
  GOOS=linux \
  GOARCH=amd64

WORKDIR /build

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build -a -ldflags "-linkmode external -extldflags '-static' -s -w" -o main main.go

WORKDIR /dist

RUN cp /build/main .

FROM scratch

COPY --from=builder /dist/main /

ENTRYPOINT ["/main"]
