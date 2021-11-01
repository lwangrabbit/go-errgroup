package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

type Server interface {
	Start() error
	Stop() error
}

type HttpServer struct {
	Addr   string
	server *http.Server
}

func NewHttpServer(addr string) *HttpServer {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		time.Sleep(5 * time.Second)
		fmt.Fprintln(writer, "I'am server"+addr)
	})
	s := &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	return &HttpServer{
		Addr:   addr,
		server: s,
	}
}

func (s *HttpServer) Start() error {
	log.Println("start server:", s.Addr)
	return s.server.ListenAndServe()
}

func (s *HttpServer) Stop() error {
	log.Println("stop server:", s.Addr)
	return s.server.Shutdown(context.TODO())
}

func main() {
	g, ctx := errgroup.WithContext(context.Background())

	servers := []*HttpServer{
		NewHttpServer(":8091"),
		NewHttpServer(":8092"),
	}
	for _, srv := range servers {
		server := srv
		g.Go(func() error {
			<-ctx.Done()
			log.Println("server: ", server.Addr, " done")
			return server.Stop()
		})
		g.Go(func() error {
			return server.Start()
		})
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	g.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				err := ctx.Err()
				log.Println("group: done")
				return err
			case <-c:
				log.Println("recv: signal")
				return errors.New("signal")
			}
		}
	})
	if err := g.Wait(); err != nil {
		log.Println("exit...")
	}
}
