package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/pukanski/loadTesting/internal/config"
	"github.com/pukanski/loadTesting/internal/display"
	"github.com/pukanski/loadTesting/internal/runner"
)

func main() {
	cfg, err := config.ParseFlags(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro: %v\n", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	fmt.Printf("Iniciando load test em %d URL(s)...\n", len(cfg.URLs))
	fmt.Printf("Concorrência: %d | Total de Requisições por URL: %d\n", cfg.Concurrency, cfg.TotalRequests)
	fmt.Printf("Pressione Ctrl+C para interromper e ver o relatório parcial.\n\n")

	reports := runner.Run(ctx, cfg)

	display.PrintReports(os.Stdout, reports)
}
