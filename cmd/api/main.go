package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/waldirborbajr/cloudstok/internal/downright"
)

var (
	port           = flag.Int("port", 3100, "port to listen on")
	timeoutSeconds = flag.Int("timeout", 10, "seconds to wait before shutting down")
)

func main() {
	flag.Parse()

	// generate a `Certificate` struct
	cert, _ := tls.LoadX509KeyPair("./certs/server.crt", "./certs/server.key")

	var err error
	defer func() {
		if err != nil {
			log.Println("exited with error: " + err.Error())
		}
	}()

	server := &GracefulServer{
		Server: &http.Server{
			Addr:    fmt.Sprintf(":%d", *port),
			Handler: downright.SlowHandler(),
			TLSConfig: &tls.Config{
				Certificates: []tls.Certificate{cert},
			},
		},
	}

	go server.WaitForExitingSignal(time.Duration(*timeoutSeconds) * time.Second)

	log.Printf("listening on port %d...", *port)
	err = server.ListenAndServeTLS()
	if err != nil {
		err = fmt.Errorf("unexpected error from ListenAndServeTLS: %w", err)
	}
	log.Println("main goroutine exited.")
}

type GracefulServer struct {
	Server           *http.Server
	shutdownFinished chan struct{}
}

func (s *GracefulServer) ListenAndServeTLS() (err error) {
	if s.shutdownFinished == nil {
		s.shutdownFinished = make(chan struct{})
	}

	err = s.Server.ListenAndServeTLS("", "")
	if err == http.ErrServerClosed {
		// expected error after calling Server.Shutdown().
		err = nil
	} else if err != nil {
		err = fmt.Errorf("unexpected error from ListenAndServe: %w", err)
		return
	}

	log.Println("waiting for shutdown finishing...")
	<-s.shutdownFinished
	log.Println("shutdown finished")

	return
}

func (s *GracefulServer) WaitForExitingSignal(timeout time.Duration) {
	var waiter = make(chan os.Signal, 1) // buffered channel
	signal.Notify(waiter, syscall.SIGTERM, syscall.SIGINT)

	// blocks here until there's a signal
	<-waiter

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	err := s.Server.Shutdown(ctx)
	if err != nil {
		log.Println("shutting down: " + err.Error())
	} else {
		log.Println("shutdown processed successfully")
		close(s.shutdownFinished)
	}
}
