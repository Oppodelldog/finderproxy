package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func startHttpProxy(ctx context.Context, localAddress, remoteAddress *net.TCPAddr) chan bool {
	done := make(chan bool)
	finalLocalAddress := fmt.Sprintf("%v:9090", localAddress.IP.String())
	origin, _ := url.Parse(fmt.Sprintf("http://%s/", remoteAddress.IP.String()))

	log.Printf("HTTP Listening: http://%v)", finalLocalAddress)
	log.Printf("HTTP Proxying: %v", origin)

	director := func(req *http.Request) {
		req.Header.Add("X-Forwarded-Host", req.Host)
		req.Header.Add("X-Origin-Host", origin.Host)
		req.URL.Scheme = "http"
		req.URL.Host = origin.Host
	}

	proxy := &httputil.ReverseProxy{Director: director}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	})
	server := &http.Server{
		Addr:              finalLocalAddress,
		Handler:           nil,
		TLSConfig:         nil,
		ReadTimeout:       0,
		ReadHeaderTimeout: 0,
		WriteTimeout:      0,
		IdleTimeout:       0,
		MaxHeaderBytes:    0,
		TLSNextProto:      nil,
		ConnState:         nil,
		ErrorLog:          nil,
	}

	go func() {
		for range ctx.Done() {
		}
		err := server.Close()
		if err != nil {
			log.Printf("error closing http server: %v", err)
		}
	}()

	go func() {
		defer close(done)
		err := server.ListenAndServe()
		if err != nil {
			log.Printf("error running http server: %v", err)
		}
		log.Println("http server closed")
	}()

	return done
}
