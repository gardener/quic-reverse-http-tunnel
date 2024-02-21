// Modifications Copyright 2024 SAP SE or an SAP affiliate company and Gardener contributors

package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"os"
	"sync"
	"time"

	quic "github.com/quic-go/quic-go"
	"k8s.io/klog/v2"

	"github.com/gardener/quic-reverse-http-tunnel/internal/pipe"
)

func main() {
	klog.Fatal(startServeer())
}

type clients struct {
	mu         sync.RWMutex
	connection []quic.Connection
	next       int
	random     *rand.Rand
}

// nextConnection returns a random connection at round-robin
func (c *clients) nextConnection() (quic.Connection, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.connection) == 0 {
		return nil, fmt.Errorf("no client connections available")
	}

	sc := c.connection[c.next]
	c.next = (c.next + 1) % len(c.connection)

	return sc, nil
}

// Start a server that echos all data on the first stream opened by the client
func startServeer() error {
	var cert, key, clientCACert, quicListener, tcpListener string

	flag.StringVar(&cert, "cert-file", "", "cert file")
	flag.StringVar(&key, "cert-key", "", "key file")
	flag.StringVar(&clientCACert, "client-ca-file", "", "client ca cert file")
	flag.StringVar(&quicListener, "listen-quic", "0.0.0.0:8888", "listen for quic")
	flag.StringVar(&tcpListener, "listen-tcp", "0.0.0.0:8443", "listen for tcp")

	klog.InitFlags(nil)
	flag.Parse()

	c := clients{
		connection: []quic.Connection{},
		random:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	conf := &quic.Config{
		HandshakeIdleTimeout:       time.Second * 2,
		MaxIdleTimeout:             time.Second * 5,
		MaxStreamReceiveWindow:     246 * (1 << 20), // 276 MB
		MaxIncomingStreams:         10000,
		MaxConnectionReceiveWindow: 500 * (1 << 20), // 512 MB,
		MaxIncomingUniStreams:      10000,
	}

	tlsCert, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		panic(err)
	}

	caPool := x509.NewCertPool()

	certBytes, err := os.ReadFile(clientCACert)
	if err != nil {
		klog.Fatalf("failed to read client CA cert file %s, got %v", clientCACert, err)
	}

	ok := caPool.AppendCertsFromPEM(certBytes)
	if !ok {
		klog.Fatalln("failed to append client CA cert to the cert pool")
	}

	tlsc := &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    caPool,
		NextProtos:   []string{"quic-echo-example"},
	}

	klog.V(1).InfoS("listening for quic connections", "address", quicListener)

	ql, err := quic.ListenAddr(quicListener, tlsc, conf)
	if err != nil {
		return err
	}

	klog.V(1).InfoS("listening for tcp connections", "address", tcpListener)

	l, err := net.Listen("tcp4", tcpListener)
	if err != nil {
		klog.Fatalf("can't listen for tcp on %s: %v", tcpListener, err)
	}

	klog.V(0).Infoln("server started")

	ctx := context.Background()

	klog.V(2).Infoln("waiting for tcp client connections")

	go func() {
		for {
			src, err := l.Accept()
			if err != nil {
				klog.ErrorS(err, "accept new tcp connection failure")

				continue
			}

			klog.V(2).InfoS("accepted TCP client connection", "remote", src.RemoteAddr())

			s, err := c.nextConnection()
			if err != nil {
				klog.ErrorS(err, "could not process tcp connection")
				src.Close()

				continue
			}

			stream, err := s.OpenStreamSync(ctx)
			if err != nil {
				klog.ErrorS(err, "cannot open stream")

				continue
			}

			klog.V(4).InfoS("opened quic stream connection", "streamID", stream.StreamID())

			go pipe.Request(src, stream)
		}
	}()

	for {
		klog.V(3).Infoln("waiting for new quic client session")

		conn, err := ql.Accept(ctx)
		if err != nil {
			klog.ErrorS(err, "unable to accept quic connection")

			continue
		}

		klog.V(2).InfoS("got a quic client session", "remote", conn.RemoteAddr())

		go func(s quic.Connection) {
			c.mu.Lock()
			c.connection = append(c.connection, s)
			c.mu.Unlock()

			<-conn.Context().Done()

			klog.V(1).InfoS("lost a client")

			c.mu.Lock()
			for i := 0; i < len(c.connection); i++ {
				if c.connection[i] != s {
					continue
				}

				c.connection[i] = c.connection[len(c.connection)-1]
				c.connection = c.connection[:len(c.connection)-1]

				if slen := len(c.connection); slen == 0 {
					c.next = 0
				} else {
					c.next = c.random.Intn(slen)
				}

				break
			}
			c.mu.Unlock()
		}(conn)
	}
}
