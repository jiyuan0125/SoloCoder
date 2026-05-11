package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/example/graceful-shutdown/pkg/graceful"
)

type server struct {
	graceful *graceful.Manager
	http     *http.Server
	mux      *http.ServeMux

	dbPool    *mockDBPool
	fileHandles []*mockFileHandle
	mu        sync.Mutex
}

type mockDBPool struct {
	connections int
	mu          sync.Mutex
}

func (p *mockDBPool) Close(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	log.Printf("Closing database connection pool with %d connections...", p.connections)

	for i := p.connections; i > 0; i-- {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(200 * time.Millisecond):
			p.connections--
			log.Printf("  Database connection %d closed", i)
		}
	}

	log.Printf("Database connection pool closed successfully")
	return nil
}

type mockFileHandle struct {
	name string
	open bool
}

func (f *mockFileHandle) Close(ctx context.Context) error {
	if !f.open {
		return nil
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(100 * time.Millisecond):
		f.open = false
		log.Printf("File handle '%s' closed", f.name)
		return nil
	}
}

func newServer() *server {
	logger := func(format string, args ...interface{}) {
		log.Printf("[GRACEFUL] "+format, args...)
	}

	gm := graceful.New(
		graceful.WithTimeout(30*time.Second),
		graceful.WithLogger(logger),
	)

	mux := http.NewServeMux()

	s := &server{
		graceful: gm,
		mux:      mux,
		dbPool: &mockDBPool{
			connections: 5,
		},
	}

	s.setupCallbacks()
	s.setupRoutes()

	return s
}

func (s *server) setupCallbacks() {
	gm := s.graceful

	gm.Register("http-server", 10, func(ctx context.Context) error {
		log.Println("Shutting down HTTP server...")

		shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		if s.http != nil {
			if err := s.http.Shutdown(shutdownCtx); err != nil {
				return fmt.Errorf("http server shutdown failed: %w", err)
			}
		}
		log.Println("HTTP server shut down successfully")
		return nil
	})

	gm.Register("database-pool", 20, func(ctx context.Context) error {
		return s.dbPool.Close(ctx)
	})

	gm.Register("file-handles", 30, func(ctx context.Context) error {
		s.mu.Lock()
		handles := make([]*mockFileHandle, len(s.fileHandles))
		copy(handles, s.fileHandles)
		s.mu.Unlock()

		for _, h := range handles {
			if err := h.Close(ctx); err != nil {
				return err
			}
		}
		return nil
	})

	gm.Register("final-cleanup", 100, func(ctx context.Context) error {
		log.Println("Performing final cleanup...")
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(50 * time.Millisecond):
			log.Println("Final cleanup completed")
			return nil
		}
	})
}

func (s *server) setupRoutes() {
	s.mux.HandleFunc("/", s.handleRoot)
	s.mux.HandleFunc("/status", s.handleStatus)
	s.mux.HandleFunc("/graceful", s.handleGraceful)
	s.mux.HandleFunc("/force", s.handleForce)
	s.mux.HandleFunc("/register", s.handleRegister)
}

func (s *server) run(addr string) error {
	s.http = &http.Server{
		Addr:    addr,
		Handler: s.mux,
	}

	log.Printf("Server starting on %s", addr)
	log.Printf("Registered callbacks:")
	for _, cb := range s.graceful.GetCallbacksInfo() {
		log.Printf("  - %s (order: %d)", cb.Name, cb.Order)
	}
	log.Println("Press Ctrl+C to trigger graceful shutdown, or press twice to force exit")

	go func() {
		if err := s.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	<-s.graceful.Done()
	log.Println("Server stopped")
	return nil
}

func main() {
	addr := ":8080"
	if len(os.Args) > 1 {
		addr = os.Args[1]
	}

	s := newServer()
	if err := s.run(addr); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
