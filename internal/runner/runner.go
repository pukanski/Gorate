package runner

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/pukanski/Gorate/internal/config"
	"github.com/pukanski/Gorate/internal/metrics"
)

func executeRequest(ctx context.Context, client *http.Client, url string, cfg *config.Config) metrics.RequestResult {
	tempoInicial := time.Now()

	var bodyReader io.Reader
	if len(cfg.Body) > 0 {
		bodyReader = bytes.NewReader(cfg.Body)
	}

	req, err := http.NewRequestWithContext(ctx, string(cfg.Method), url, bodyReader)
	if err != nil {
		return metrics.RequestResult{
			URL:        url,
			Duration:   time.Since(tempoInicial),
			StatusCode: 0,
			IsTimeout:  false,
			Err:        err,
		}
	}

	for key, value := range cfg.HeadersCustom {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	duracao := time.Since(tempoInicial)

	if err != nil {
		isTimeout := false
		var netErr net.Error

		if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &netErr) && netErr.Timeout()) {
			isTimeout = true
		}

		statusCode := 0
		if resp != nil {
			statusCode = resp.StatusCode
		}

		return metrics.RequestResult{
			URL:        url,
			Duration:   duracao,
			StatusCode: statusCode,
			IsTimeout:  isTimeout,
			Err:        err,
		}
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	return metrics.RequestResult{
		URL:        url,
		Duration:   duracao,
		StatusCode: resp.StatusCode,
		IsTimeout:  false,
		Err:        err,
	}
}

func runSingleURL(ctx context.Context, client *http.Client, url string, cfg *config.Config) metrics.Report {
	jobs := make(chan string, cfg.TotalRequests)
	results := make(chan metrics.RequestResult, cfg.TotalRequests)

	var wg sync.WaitGroup

	inicioUrl := time.Now()

	for i := 0; i < cfg.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range jobs {
				if ctx.Err() != nil {
					return
				}

				res := executeRequest(ctx, client, url, cfg)
				if errors.Is(res.Err, context.Canceled) {
					return
				}

				results <- res
			}
		}()
	}

	go func() {
		defer close(jobs)

		var limiter *time.Ticker
		var limiterChan <-chan time.Time

		if cfg.RateLimit > 0 {
			intervalo := time.Second / time.Duration(cfg.RateLimit)
			if intervalo > 0 {
				limiter = time.NewTicker(intervalo)
				defer limiter.Stop()
				limiterChan = limiter.C
			}
		}

		for i := 0; i < cfg.TotalRequests; i++ {
			if limiterChan != nil {
				select {
				case <-ctx.Done():
					return
				case <-limiterChan:
				}
			}

			select {
			case <-ctx.Done():
				return
			case jobs <- url:
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	var urlResults []metrics.RequestResult
	for res := range results {
		urlResults = append(urlResults, res)
	}

	tempoTotal := time.Since(inicioUrl)

	return metrics.GenerateReport(url, urlResults, tempoTotal)
}

func Run(ctx context.Context, cfg *config.Config) []metrics.Report {
	client := &http.Client{
		Timeout: cfg.Timeout,
		Transport: &http.Transport{
			MaxIdleConnsPerHost: cfg.Concurrency,
		},
	}

	reports := make([]metrics.Report, len(cfg.URLs))

	var wg sync.WaitGroup

	for i, u := range cfg.URLs {
		wg.Add(1)
		go func(index int, url string) {
			defer wg.Done()
			reports[index] = runSingleURL(ctx, client, url, cfg)
		}(i, u)
	}

	wg.Wait()

	return reports
}
