package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/BenDowswell/kvstore1/server"
	"github.com/BenDowswell/kvstore1/store"
)

func main() {

	addr := flag.String("addr", ":8080", "HTTP network address")
	flag.Parse()
	repo := store.NewMemoryStore()
	svc := store.NewKVService(repo)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	err := server.StartServer(ctx, *addr, svc)
	if err != nil {
		slog.Error("server failed to start", "error", err)
		os.Exit(1)
	}
}
