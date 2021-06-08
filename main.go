package main

import (
	"context"
	"fmt"
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
	return s.server.ListenAndServe()
}

func (s *HttpServer) Stop() error {
	return s.server.Shutdown(context.TODO())
}

func main() {
	pctx, cancel := context.WithCancel(context.TODO())
	g, ctx := errgroup.WithContext(pctx)

	servers := []*HttpServer{
		NewHttpServer(":8091"),
		NewHttpServer(":8092"),
	}
	for _, srv := range servers {
		server := srv
		g.Go(func() error {
			<-ctx.Done()
			fmt.Println(time.Now(), ": done")
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
				fmt.Println(time.Now(), ":for done")
				return err
			case <-c:
				fmt.Println(time.Now(), ":cancel")
				cancel()
			}
		}
	})
	if err := g.Wait(); err != nil {
		fmt.Println("exit...")
	}
}
