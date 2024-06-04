# Modifications Copyright 2024 SAP SE or an SAP affiliate company and Gardener contributors

############# builder
FROM golang:1.22.4 AS builder
ARG GOPROXY=https://proxy.golang.org,direct
ENV GOPROXY=$GOPROXY
WORKDIR /go/src/github.com/gardener/quic-reverse-http-tunnel
COPY . .
RUN make install


############# distroless-static
FROM gcr.io/distroless/static-debian12:nonroot as distroless-static

############# server
FROM distroless-static AS quic-server
COPY --from=builder /go/bin/server /server
WORKDIR /
ENTRYPOINT ["/server"]

############# client
FROM distroless-static AS quic-client
COPY --from=builder /go/bin/client /client
WORKDIR /
ENTRYPOINT ["/client"]

############# client-tcp
FROM distroless-static AS quic-client-tcp
COPY --from=builder /go/bin/client-tcp /client-tcp
WORKDIR /
ENTRYPOINT ["/client-tcp"]
