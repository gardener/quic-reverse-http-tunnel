# Modifications Copyright 2024 SAP SE or an SAP affiliate company and Gardener contributors

#############      builder       #############
FROM golang:1.21.7 AS builder
ARG GOPROXY=https://proxy.golang.org,direct
ENV GOPROXY=$GOPROXY
WORKDIR /go/src/github.com/plkokanov/quic-reverse-http-tunnel
COPY . .
RUN make install

#############      server     #############
FROM alpine:3.18.4 AS server
RUN apk add --update tzdata
COPY --from=builder /go/bin/server /server
WORKDIR /
ENTRYPOINT ["/server"]

############# client #############
FROM alpine:3.18.4 AS client
RUN apk add --update tzdata
COPY --from=builder /go/bin/client /client
WORKDIR /
ENTRYPOINT ["/client"]

############# client-tcp #############
FROM alpine:3.18.4 AS client-tcp
RUN apk add --update tzdata
COPY --from=builder /go/bin/client-tcp /client-tcp
WORKDIR /
ENTRYPOINT ["/client-tcp"]
