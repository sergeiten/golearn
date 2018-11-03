FROM golang:1.9 as builder

RUN mkdir -p /go/src/github.com/sergeiten/golearn
WORKDIR /go/src/github.com/sergeiten/golearn
COPY . .

RUN CGO_ENABLED=0 go build -a -installsuffix cgo -o app cmd/golearn/main.go

FROM alpine:latest
RUN set -ex && apk add --no-cache ca-certificates
WORKDIR /
COPY --from=builder /go/src/github.com/sergeiten/golearn/app .
COPY --from=builder /go/src/github.com/sergeiten/golearn/lang.json .
# CMD ["/app", "-port="${PORT}]
CMD ./app --port=$CONTAINER_PORT
