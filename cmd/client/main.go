// Modifications Copyright 2024 SAP SE or an SAP affiliate company and Gardener contributors

package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	quic "github.com/quic-go/quic-go"
	"k8s.io/klog/v2"
)

// We start a server echoing data on the first stream the client opens,
// then connect with a client, send the message, and wait for its receipt.
func main() {
	err := clientMain()
	if err != nil {
		panic(err)
	}
}

const (
	ListenerCloseCode quic.ApplicationErrorCode = 100
)

var (
	_ net.Listener = &listener{}
	_ net.Conn     = &conn{}
)

// listener implements net.Listener
type listener struct {
	connection quic.Connection
	ctx        context.Context
}

func (h *listener) Accept() (net.Conn, error) {
	s, err := h.connection.AcceptStream(h.ctx)
	if err != nil {
		return nil, err
	}

	return &conn{
		Stream: s,
		local:  h.connection.LocalAddr(),
		remote: h.connection.RemoteAddr(),
	}, nil
}

func (h *listener) Close() error {
	return h.connection.CloseWithError(ListenerCloseCode, "die")
}

func (h *listener) Addr() net.Addr {
	return nil
}

// conn implements net.Conn.
type conn struct {
	quic.Stream
	local, remote net.Addr
}

func (h *conn) LocalAddr() net.Addr {
	return h.local
}

func (h *conn) RemoteAddr() net.Addr {
	return h.remote
}

type connectServer struct {
	connectResponse []byte
}

func (c *connectServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	klog.V(2).InfoS("Received", "host", r.Host, "method", r.Method, "user-agent", r.UserAgent())

	if r.Method != http.MethodConnect {
		http.Error(w, "this proxy only supports CONNECT passthrough", http.StatusMethodNotAllowed)

		return
	}

	// Connect to Remote.
	dst, err := net.Dial("tcp", r.RequestURI)
	if err != nil {
		klog.ErrorS(err, "could not dial requested host", "host", r.RequestURI)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer dst.Close()

	// Upon success, we respond a 200 status code to client.
	if _, err := w.Write(c.connectResponse); err != nil {
		klog.ErrorS(err, "could not write 200 response")
		return
	}

	// Now, Hijack the writer to get the underlying net.Conn.
	// Which can be either *tcp.Conn, for HTTP, or *tls.Conn, for HTTPS.
	src, bio, err := w.(http.Hijacker).Hijack()
	if err != nil {
		klog.ErrorS(err, "could not hijack connection", "host", r.RequestURI)

		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
	defer src.Close()

	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()

		// Copy unprocessed buffered data from the client to dst so we can use src directly.
		if n := bio.Reader.Buffered(); n > 0 {
			n64, err := io.CopyN(dst, bio, int64(n))
			if n64 != int64(n) || err != nil {
				klog.ErrorS(err, "io.Copy failure", "n64", n64, "host", r.RequestURI)

				return
			}
		}

		// src -> dst
		if _, err := io.Copy(dst, src); err != nil {
			klog.ErrorS(err, "cant copy from source to destination", "host", r.RequestURI)
		}
	}()

	go func() {
		defer wg.Done()

		// dst -> src
		if _, err := io.Copy(src, dst); err != nil {
			klog.ErrorS(err, "cant copy from destination to source", "host", r.RequestURI)
		}
	}()

	wg.Wait()
}

func clientMain() error {
	var ca, cert, key, server string

	flag.StringVar(&ca, "ca-file", "", "ca file")
	flag.StringVar(&server, "server", "127.0.0.1:9999", "host:port of the quic server")
	flag.StringVar(&cert, "cert-file", "", "client cert file")
	flag.StringVar(&key, "cert-key", "", "client key file")

	klog.InitFlags(nil)

	flag.Parse()

	data, err := os.ReadFile(ca)
	if err != nil {
		return fmt.Errorf("could not read certificate authority: %s", err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(data) {
		return fmt.Errorf("could not append certificate data")
	}

	tlsCert, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		return fmt.Errorf("could not read client certificates: %s", err)
	}

	tlsConf := &tls.Config{
		ServerName:   "quic-tunnel-server",
		Certificates: []tls.Certificate{tlsCert},
		RootCAs:      certPool,
		NextProtos:   []string{"quic-echo-example"},
	}

	conf := &quic.Config{
		KeepAlivePeriod:            time.Second * 2,
		HandshakeIdleTimeout:       time.Second * 2,
		MaxIdleTimeout:             time.Second * 5,
		MaxStreamReceiveWindow:     246 * (1 << 20), // 276 MB
		MaxIncomingStreams:         10000,
		MaxConnectionReceiveWindow: 500 * (1 << 20), // 512 MB,
		MaxIncomingUniStreams:      10000,
	}

	ctx := context.Background()

	for {
		klog.V(2).InfoS("dialing quic server", "remote", server)

		session, err := quic.DialAddr(ctx, server, tlsConf, conf)
		if err != nil {
			// TODO this needs backoff
			klog.ErrorS(err, "could not dial quic server")

			continue
		}

		go func() {
			<-session.Context().Done()
			klog.V(2).Infoln("session closed.")
		}()

		klog.V(2).Infoln("starting http server")

		err = http.Serve(&listener{connection: session, ctx: ctx}, &connectServer{
			connectResponse: []byte("HTTP/1.1 200 OK\r\n\r\n"),
		})
		if err != nil {
			klog.ErrorS(err, "failure on http serving")
		}
	}
}
