package main

import (
	"context"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

var flagLocalAddr = flag.String("l", "localhost:9999", "local address")
var flagRemoteAddr = flag.String("r", "localhost:80", "remote address")

var remoteAddress, localAddress *net.TCPAddr

func init() {
	var err error

	flag.Parse()

	log.Printf("resolving local address: %s", *flagLocalAddr)
	localAddress, err = net.ResolveTCPAddr("tcp4", *flagLocalAddr)
	if err != nil {
		panic(err)
	}
	log.Printf("resolved local address: %s", localAddress.String())

	log.Printf("resolving remote address: %s", *flagRemoteAddr)
	remoteAddress, err = net.ResolveTCPAddr("tcp4", *flagRemoteAddr)
	if err != nil {
		panic(err)
	}
	log.Printf("resolved remote address: %s", remoteAddress.String())
}

func main() {
	ctx := newSignalContextWithSignals(os.Interrupt, syscall.SIGTERM)

	tcpProxyDone := startTcpProxy(ctx, localAddress, remoteAddress)
	httpProxyDone := startHttpProxy(ctx, localAddress, remoteAddress)

	<-ctx.Done()
	<-tcpProxyDone
	<-httpProxyDone
}

func newSignalContextWithSignals(signals ...os.Signal) context.Context {

	ctx, cancelFunc := context.WithCancel(context.Background())

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, signals...)

		for {
			select {
			case <-c:
				cancelFunc()
				return
			}
		}
	}()

	return ctx
}
