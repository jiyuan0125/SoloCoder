package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"unixsocket-fdpass/api"
	"unixsocket-fdpass/fdpass"
)

var (
	listenAddr   = flag.String("addr", ":8402", "listen address")
	restartMu    sync.Mutex
	isRestarting bool
	startTime    = time.Now()
	hrm          *fdpass.HotReloadManager
)

func main() {
	flag.Parse()

	hrm = fdpass.NewHotReloadManager(fdpass.HotReloadConfig{
		ShutdownTimeout: 30 * time.Second,
	})

	var ln net.Listener
	var err error

	if !hrm.IsPrimary() {
		ln, err = hrm.ChildHandshake()
		if err != nil {
			log.Printf("child handshake failed: %v, falling back to normal listen", err)
			ln, err = hrm.Start(*listenAddr)
			if err != nil {
				log.Fatalf("failed to start listener: %v", err)
			}
		}
	} else {
		ln, err = hrm.Start(*listenAddr)
		if err != nil {
			log.Fatalf("failed to start listener: %v", err)
		}
	}

	log.Printf("server started, pid=%d, addr=%s, isPrimary=%v",
		os.Getpid(), ln.Addr(), hrm.IsPrimary())

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/status", statusHandler)
	mux.HandleFunc("/echo", echoHandler)
	mux.HandleFunc("/restart", restartHandler)
	mux.HandleFunc("/history", historyHandler)

	srv := &http.Server{
		Handler: mux,
	}

	go handleSignals(srv, ln)

	if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
		log.Fatalf("serve error: %v", err)
	}

	log.Printf("server exited, pid=%d", os.Getpid())
}

func handleSignals(srv *http.Server, ln net.Listener) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR2)

	for sig := range sigCh {
		switch sig {
		case syscall.SIGUSR2:
			go doRestart()
		case syscall.SIGINT, syscall.SIGTERM:
			log.Printf("received shutdown signal, pid=%d", os.Getpid())
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			_ = srv.Shutdown(ctx)
			return
		}
	}
}

func doRestart() {
	restartMu.Lock()
	if isRestarting {
		restartMu.Unlock()
		log.Println("restart already in progress")
		return
	}
	isRestarting = true
	restartMu.Unlock()

	log.Printf("starting hot restart, old pid=%d", os.Getpid())

	newPID, err := hrm.Restart()
	if err != nil {
		log.Printf("restart failed: %v", err)
		restartMu.Lock()
		isRestarting = false
		restartMu.Unlock()
		return
	}

	log.Printf("restart successful, new pid=%d, old pid=%d exiting", newPID, os.Getpid())
	time.Sleep(100 * time.Millisecond)
	os.Exit(0)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	resp := api.HealthResponse{
		Status:    "ok",
		PID:       os.Getpid(),
		Timestamp: time.Now().Unix(),
	}
	writeJSON(w, http.StatusOK, resp)
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	resp := api.StatusResponse{
		PID:          os.Getpid(),
		ListenAddr:   *listenAddr,
		IsPrimary:    hrm.IsPrimary(),
		RestartCount: hrm.RestartCount(),
		StartTime:    startTime.Unix(),
	}
	writeJSON(w, http.StatusOK, resp)
}

func echoHandler(w http.ResponseWriter, r *http.Request) {
	msg := r.URL.Query().Get("msg")
	if msg == "" {
		msg = "hello"
	}
	resp := api.EchoResponse{
		PID:       os.Getpid(),
		Message:   fmt.Sprintf("echo: %s", msg),
		Timestamp: time.Now().Unix(),
	}
	writeJSON(w, http.StatusOK, resp)
}

func restartHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	restartMu.Lock()
	if isRestarting {
		restartMu.Unlock()
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte(`{"status":"restarting"}`))
		return
	}
	isRestarting = true
	restartMu.Unlock()

	go doRestart()
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"status":"accepted"}`))
}

func historyHandler(w http.ResponseWriter, r *http.Request) {
	records := hrm.RestartHistory()
	history := make([]api.RestartHistory, 0, len(records))
	for _, rec := range records {
		history = append(history, api.RestartHistory{
			OldPID:     rec.OldPID,
			NewPID:     rec.NewPID,
			Timestamp:  rec.Timestamp.Unix(),
			Successful: rec.Successful,
		})
	}
	resp := api.RestartHistoryResponse{
		History: history,
	}
	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
