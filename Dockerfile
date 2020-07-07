FROM golang:1.14.4-alpine

WORKDIR /build

COPY  . .

RUN go build -o stress stress.go

FROM alpine:3.12.0

ENV LISTEN_ADDRESS 0.0.0.0
ENV LISTEN_PORT 8080

WORKDIR /opt

COPY --from=0 /build/stress stress

ENTRYPOINT ["/opt/stress"]
