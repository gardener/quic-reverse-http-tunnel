# quic-reverse-http-tunnel

[![REUSE status](https://api.reuse.software/badge/github.com/gardener/quic-reverse-http-tunnel)](https://api.reuse.software/info/github.com/gardener/quic-reverse-http-tunnel)

## What it does

It's a reverse HTTP Tunnel using QUIC:

```text
K8S apiserver / curl --- TCP ----> [proxy-server] ---- QUIC ----> [proxy-agent]---TCP--> [kubelet]
```

1. the proxy-server listens for `tcp` (no HTTP server running) and `quic`.
1. The proxy-agent talks to the server and opens a `quic` session.
1. It starts a HTTP tunnel server that listens on that session for new streams.
1. When the API server / curl talks to the proxy-server, it creates a new `quic` stream and sends the data to the proxy-agent.
1. The HTTP server in the proxy-agent that listens on new quic streams accepts the stream, opens TCP connection to the requested host (from the CONNECT) and pipes the data back.

The proxy can also run as a simple passthrough proxy via `client-tcp`
## Building and running

Run the server:

```console
$ make start-server
2020/11/01 02:11:39 quick listener on 0.0.0.0:8888
2020/11/01 02:11:39 tcp listener on 0.0.0.0:10443
2020/11/01 02:11:39 waiting for new quic client session
2020/11/01 02:11:39 waiting for tcp client connections
```

in another terminal run the client:

```console
$ make start-client
2020/11/01 02:13:31 dialing quic server...
2020/11/01 02:13:31 starting http server
```

and in third try to access it:

```console
curl -p --proxy localhost:10443 http://www.example.com
```

If you want to test the passthrough proxy instead:

```console
$ make start-client-tcp
2020/11/25 12:07:07 dialing quic server...
2020/11/25 12:07:07 connected to quic server
```

## Docker images

Docker images are available at:

- `ghcr.io/gardener/quic-reverse-http-tunnel/quic-server:latest`
- `ghcr.io/gardener/quic-reverse-http-tunnel/quic-client:latest`
- `ghcr.io/gardener/quic-reverse-http-tunnel/quic-client-tcp:latest`

If you want to use a specific version tag, check the [`tags` of the gardener/quic-reverse-http-tunnel](https://github.com/gardener/quic-reverse-http-tunnel/tags) repository.

A [github action](./.github/workflows/image-quic-reverse-http-tunnel.yaml) takes care of building and pushing new images to `ghcr.io` when a new github tag is created.

To build the images locally, you can use the `make docker-images` command:
```console
REGISTRY=<your-registry> IMAGE_TAG=<your-image-tag>  make docker-images
```

**Note** that if you do not specify the `REGISTRY` and `IMAGE_TAG` variables, the default ones from the [`Makefile`](./Makefile) will be used.