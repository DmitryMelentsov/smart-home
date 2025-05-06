package http

import (
	"context"
	"fmt"
	"homework/internal/usecase"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	host   string
	port   uint16
	router *gin.Engine
	ws     *WebSocketHandler
}

type UseCases struct {
	Event  *usecase.Event
	Sensor *usecase.Sensor
	User   *usecase.User
}

func NewServer(useCases UseCases, options ...func(*Server)) *Server {
	r := gin.Default()
	tmp := NewWebSocketHandler(useCases)
	setupRouter(r, useCases, tmp)

	s := &Server{router: r, host: "localhost", port: 8080, ws: tmp}
	for _, o := range options {
		o(s)
	}

	return s
}

func WithHost(host string) func(*Server) {
	return func(s *Server) {
		s.host = host
	}
}

func WithPort(port uint16) func(*Server) {
	return func(s *Server) {
		s.port = port
	}
}

func (s *Server) Run(ctx context.Context) error {
	eg, appCtx := errgroup.WithContext(ctx)
	sigQuit := make(chan os.Signal, 1)
	signal.Notify(sigQuit, syscall.SIGINT, syscall.SIGTERM)
	eg.Go(func() error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case s := <-sigQuit:
			return fmt.Errorf("signal: %v", s)
		}
	})
	eg.Go(func() error {
		<-appCtx.Done()
		_ = s.ws.Shutdown()
		return fmt.Errorf("done")
	})
	eg.Go(func() error {
		return s.router.Run(fmt.Sprintf("%s:%d", s.host, s.port))
	})

	if err := eg.Wait(); err != nil {
		return fmt.Errorf("server stopped with error: %w", err)
	}
	return nil
}
