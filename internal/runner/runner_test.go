package runner

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pukanski/loadTesting/internal/config"
)

func TestRun_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	cfg := &config.Config{
		URLs:          []string{server.URL},
		TotalRequests: 10,
		Concurrency:   5,
		Method:        config.MethodGet,
		Timeout:       2 * time.Second,
		RateLimit:     0,
	}

	reports := Run(context.Background(), cfg)

	if len(reports) != 1 {
		t.Fatalf("Esperava 1 relatório, obteve %d", len(reports))
	}

	r := reports[0]
	if r.TotalRequests != 10 {
		t.Errorf("Esperava 10 requisições totais, obteve %d", r.TotalRequests)
	}
	if r.ErrorRate != 0 {
		t.Errorf("Esperava taxa de erro 0, obteve %.2f%%", r.ErrorRate)
	}
	if r.Status4xx5xx != 0 {
		t.Errorf("Esperava zero erros HTTP, obteve %d", r.Status4xx5xx)
	}
}

func TestRun_RateLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &config.Config{
		URLs:          []string{server.URL},
		TotalRequests: 5,
		Concurrency:   5,
		Method:        config.MethodGet,
		Timeout:       2 * time.Second,
		RateLimit:     10,
	}

	start := time.Now()
	reports := Run(context.Background(), cfg)
	elapsed := time.Since(start)

	if len(reports) != 1 {
		t.Fatalf("Esperava 1 relatório, obteve %d", len(reports))
	}

	if reports[0].TotalRequests != 5 {
		t.Errorf("Esperava 5 requisições totais, obteve %d", reports[0].TotalRequests)
	}

	if elapsed < 450*time.Millisecond {
		t.Errorf("Rate limit NÃO foi respeitado. Duração muito rápida: %v", elapsed)
	}
}

func TestRun_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &config.Config{
		URLs:          []string{server.URL},
		TotalRequests: 3,
		Concurrency:   3,
		Method:        config.MethodGet,
		Timeout:       20 * time.Millisecond,
		RateLimit:     0,
	}

	reports := Run(context.Background(), cfg)

	if len(reports) != 1 {
		t.Fatalf("Esperava 1 relatório, obteve %d", len(reports))
	}

	r := reports[0]
	if r.TotalRequests != 3 {
		t.Errorf("Esperava 3 requisições totais, obteve %d", r.TotalRequests)
	}

	if r.Timeouts != 3 {
		t.Errorf("Esperava 3 timeouts contabilizados, obteve %d", r.Timeouts)
	}
	if r.ErrorRate != 100 {
		t.Errorf("Esperava taxa de erro de 100%%, obteve %.2f%%", r.ErrorRate)
	}
}

func TestRun_ContextCanceled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &config.Config{
		URLs:          []string{server.URL},
		TotalRequests: 10,
		Concurrency:   1,
		Method:        config.MethodGet,
		Timeout:       500 * time.Millisecond,
		RateLimit:     0,
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(30 * time.Millisecond)
		cancel()
	}()

	reports := Run(ctx, cfg)

	if reports[0].TotalRequests == 10 {
		t.Errorf("Contexto foi cancelado mas o Runner executou todas as 10 requisições")
	}
}
