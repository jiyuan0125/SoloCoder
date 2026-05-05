package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go-db-dumper/proto"
	"go-db-dumper/server"
)

func main() {
	port := flag.Int("port", proto.DefaultPort, "Server port")
	host := flag.String("host", proto.DefaultHost, "Server host")
	flag.Parse()

	addr := fmt.Sprintf("%s:%d", *host, *port)

	srv := server.NewServer(proto.DefaultConverterConfig())

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.Start(addr); err != nil {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			os.Exit(1)
		}
	}()

	<-sigCh
	fmt.Println("\nShutting down server...")
	srv.Stop()
}
