FROM golang:1.11 as builder

RUN mkdir -p /go/src/github.com/sergeiten/golearn
WORKDIR /go/src/github.com/sergeiten/golearn
COPY . .

RUN CGO_ENABLED=0 go build -a -installsuffix cgo -o app cmd/fetch/main.go

FROM alpine:latest
RUN set -ex && apk add --no-cache ca-certificates
WORKDIR /
COPY --from=builder /go/src/github.com/sergeiten/golearn/app .
COPY --from=builder /go/src/github.com/sergeiten/golearn/credentials.json .
COPY --from=builder /go/src/github.com/sergeiten/golearn/token.json .
CMD ./app
