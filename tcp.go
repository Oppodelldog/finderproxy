package main

import (
	"context"
	"io"
	"log"
	"net"
)

const buffer = 4096

func startTcpProxy(ctx context.Context, localAddress, remoteAddress *net.TCPAddr) chan bool {
	done := make(chan bool)
	log.Printf("TCP Listening: %v", localAddress.String())
	log.Printf("TCP Proxying: %v", remoteAddress.String())

	listener, err := net.ListenTCP("tcp", localAddress)
	if err != nil {
		panic(err)
	}

	newConnections := make(chan *net.TCPConn)

	go handleConn(ctx, newConnections)
	go func() {
		defer close(done)
		for {
			select {
			case <-ctx.Done():
				log.Printf("stopping listener: %v", listener.Close())
				return
			}
		}
	}()

	go func() {
		defer close(done)
		for {
			log.Print("waiting for new TCP connection")
			conn, err := listener.AcceptTCP()
			if err != nil {
				log.Printf("error running TCP listener: %v", err)
			}
			log.Print("accepted new TCP connection")
			newConnections <- conn
		}
	}()

	return done
}

func newProxyChannel(ctx context.Context, cancelFunc context.CancelFunc, r io.ReadCloser, w io.WriteCloser) {
	dataPipe := make(chan []byte)
	maxBytes := 0
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Printf("done - closing reader: %v", r.Close())
				return
			default:
				data := make([]byte, buffer)
				n, err := r.Read(data)
				if err != nil {
					log.Printf("aborting connection - error reading: %v\n", err)
					cancelFunc()
				}
				if n > 0 {
					if n > maxBytes {
						maxBytes = n
						log.Printf("max %v bytes", maxBytes)
					}

					dataPipe <- data[:n]
				}
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Printf("done - closing writer: %v", w.Close())
				return
			case data := <-dataPipe:
				_, err := w.Write(data)
				if err != nil {
					log.Printf("aborting connection - error writing: %v", err)
					cancelFunc()
				}
			}
		}
	}()
}

func newProxyConnection(ctx context.Context, clientConnection *net.TCPConn, targetConnection *net.TCPConn) {
	proxyCtx, cancelFunc := context.WithCancel(ctx)
	newProxyChannel(proxyCtx, cancelFunc, clientConnection, targetConnection)
	newProxyChannel(proxyCtx, cancelFunc, targetConnection, clientConnection)
}

func handleConn(ctx context.Context, newConnections <-chan *net.TCPConn) {
	for clientConnection := range newConnections {
		log.Printf("starting new proxy connection for: %v", clientConnection.RemoteAddr().String())
		targetConnection, err := net.DialTCP("tcp", nil, remoteAddress)
		if err != nil {
			panic(err)
		}
		newProxyConnection(ctx, clientConnection, targetConnection)
	}
}
