package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

func main() {
	if err := run(); err != nil {
		slog.Error("demo stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	n, err := strconv.ParseUint(port, 10, 16)
	if err != nil || n == 0 {
		return errors.New("PORT must be an integer between 1 and 65535")
	}

	server, err := newServer()
	if err != nil {
		return fmt.Errorf("create server: %w", err)
	}
	listener, err := net.Listen("tcp", net.JoinHostPort("", port))
	if err != nil {
		return err
	}
	return server.Serve(ctx, listener)
}
