package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	server   http.Server
	mux      *http.ServeMux
	errgroup errgroup.Group
	sigch    chan os.Signal
	err      error
	mu       sync.Mutex
}

func (serv *Server) Start(addr string, h http.Handler, sigset []os.Signal) error {
	serv.sigch = make(chan os.Signal, 1)
	if len(sigset) > 0 {
		signal.Notify(serv.sigch, sigset...)
	}

	handler := h
	if handler == nil {
		handler = serv.mux
	}
	serv.server = http.Server{Addr: addr, Handler: handler}

	serv.errgroup.Go(func() error {
		fmt.Println("Server start...")
		err := serv.server.ListenAndServe()
		if err == http.ErrServerClosed && serv.err != nil {
			return serv.err
		}
		serv.err = err
		return err
	})

	serv.errgroup.Go(func() error {
		sig := <-serv.sigch
		if sig != nil {
			//because Stop()->Shutdown() will set errgroup.err first, so here save the sig error
			serv.err = errors.Errorf("Receive Signal %s\n", sig.String())
			serv.Stop()
		}
		return nil
	})

	return serv.errgroup.Wait()

}

func (serv *Server) Stop() error {
	close(serv.sigch)
	return serv.server.Shutdown(context.TODO()) //nil

}

func (serv *Server) HandleFunc(pattern string, handlerFunc http.HandlerFunc) {
	if serv.mux == nil {
		serv.mu.Lock()
		if serv.mux == nil {
			serv.mux = http.NewServeMux()
		}
		serv.mu.Unlock()
	}
	serv.mux.HandleFunc(pattern, handlerFunc)
}

func (serv *Server) Error() error {
	return serv.err
}

func main() {
	serv := &Server{}
	sigs := []os.Signal{
		syscall.SIGINT,
		syscall.SIGTERM,
	}

	serv.HandleFunc("/index",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "Hello World")
		})

	go func() { //killed because of timeout or signal
		time.Sleep(10 * time.Second)
		serv.Stop()
	}()

	err := serv.Start("localhost:8080", nil, sigs)
	if err != nil {
		fmt.Printf("Error happened when server running: %v\n", err)
	}

	time.Sleep(1 * time.Second)
}
