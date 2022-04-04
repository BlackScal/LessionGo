package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	http.Server
	Name    string
	running int32
}

func (srv *Server) HandleFunc(pattern string, handlerFunc http.HandlerFunc) {
	if srv.Server.Handler == nil {
		srv.Server.Handler = http.NewServeMux()
	}
	mux := srv.Server.Handler.(*http.ServeMux)
	mux.HandleFunc(pattern, handlerFunc)
}

func (srv *Server) Shutdown() error {
	var err error
	if atomic.SwapInt32(&srv.running, 0) == 1 {
		err = srv.Server.Shutdown(context.TODO())
	}
	return err
}

func (srv *Server) ListenAndServe() error {
	if atomic.SwapInt32(&srv.running, 1) == 1 {
		return nil
	}
	defer func() { atomic.SwapInt32(&srv.running, 0) }()

	return srv.Server.ListenAndServe()
}

type ServerGroup struct {
	servers  []*Server
	errgroup errgroup.Group
	sigch    chan os.Signal
	err      error
	running  int32
}

func (sg *ServerGroup) ListenAndServe(servers []*Server, sigset []os.Signal) error {
	if len(servers) <= 0 {
		return errors.Errorf("Empty server list.")
	}

	if atomic.SwapInt32(&sg.running, 0) == 1 {
		return errors.Errorf("Server Group is running.")
	}

	sg.sigch = make(chan os.Signal)
	if len(sigset) > 0 {
		signal.Notify(sg.sigch, sigset...)
	}

	sg.servers = []*Server{}
	atomic.StoreInt32(&sg.running, 1)
	for _, server := range servers {
		server := server
		sg.servers = append(sg.servers, server)

		sg.errgroup.Go(func() error {
			fmt.Printf("Server %s start...\n", server.Name)
			err := server.ListenAndServe()
			err = errors.Wrapf(err, "Server %s stopped", server.Name)
			if atomic.LoadInt32(&sg.running) == 1 {
				sg.err = err
				sg.Stop()
			}
			fmt.Printf("Server %s stop...\n", server.Name)
			return err
		})
	}

	sg.errgroup.Go(func() error {
		var err error
		sigrecv := <-sg.sigch
		if sigrecv != nil {
			err = errors.Errorf("Receive Signal %s\n", sigrecv.String())
			sg.err = err
		}
		sg.Stop()
		return err
	})

	sg.errgroup.Wait()
	return sg.err

}

func (sg *ServerGroup) Stop() error {
	if atomic.SwapInt32(&sg.running, 0) == 0 {
		return nil
	}
	close(sg.sigch)
	sg.sigch = nil
	for _, server := range sg.servers {
		server.Shutdown()
	}

	return nil

}

func (serv *ServerGroup) Sig(sig os.Signal) {
	serv.sigch <- sig
}

func (serv *ServerGroup) Error() error {
	return serv.err
}

func main() {
	var servers []*Server
	var abortServer *Server

	basePort := 8080
	for num := 0; num < 10; num++ {
		num := num
		srv := &Server{}
		srv.Name = fmt.Sprintf("Serv%03d", num)
		srv.Server.Addr = fmt.Sprintf("localhost:%d", basePort+num)
		srv.HandleFunc(
			"/index",
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintf(w, "Hello, This is Server%03d!\n", num)
			},
		)
		servers = append(servers, srv)
		if num == 1 {
			abortServer = srv
			_ = abortServer
		}
	}

	sg := &ServerGroup{}
	sigs := []os.Signal{
		syscall.SIGINT,
		syscall.SIGTERM,
	}

	go func() {
		time.Sleep(10 * time.Second)
		//sg.Stop()
		//abortServer.Shutdown()
		sg.Sig(syscall.SIGINT)
	}()

	err := sg.ListenAndServe(servers, sigs)
	if err != nil {
		fmt.Printf("Error happened when server running: %v\n", err)
	}

	time.Sleep(1 * time.Second)
}
