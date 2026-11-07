package metrics

import (
	"math"
	"slices"
	"time"
)

type RequestResult struct {
	URL        string
	Duration   time.Duration
	StatusCode int
	IsTimeout  bool
	Err        error
}

type Report struct {
	URL           string
	TotalRequests int
	RPS           float64
	ErrorRate     float64
	Status4xx5xx  int
	Timeouts      int
	P50           time.Duration
	P95           time.Duration
	P99           time.Duration
	Min           time.Duration
	Max           time.Duration
}

func CalculatePercentile(latenciasOrdenadas []time.Duration, p float64) time.Duration {
	if len(latenciasOrdenadas) == 0 {
		return 0
	}

	percentileIndex := int(math.Ceil(p*float64(len(latenciasOrdenadas)))) - 1

	if percentileIndex < 0 {
		percentileIndex = 0
	}
	if percentileIndex >= len(latenciasOrdenadas) {
		percentileIndex = len(latenciasOrdenadas) - 1
	}

	return latenciasOrdenadas[percentileIndex]
}

func GenerateReport(url string, results []RequestResult, totalElapsed time.Duration) Report {
	if len(results) == 0 || totalElapsed <= 0 {
		return Report{URL: url}
	}

	countErrors := 0
	count4xx5xx := 0
	countTimeouts := 0
	var latencias []time.Duration

	for _, req := range results {
		isHTTPError := req.StatusCode >= 400 && req.StatusCode < 600
		if isHTTPError {
			count4xx5xx++
		}
		if req.IsTimeout {
			countTimeouts++
		}
		if req.Err != nil || isHTTPError {
			countErrors++
		}
		if req.Err == nil && !req.IsTimeout {
			latencias = append(latencias, req.Duration)
		}
	}

	var min, max, p50, p95, p99 time.Duration

	if len(latencias) > 0 {
		slices.Sort(latencias)
		min = latencias[0]
		max = latencias[len(latencias)-1]

		p50 = CalculatePercentile(latencias, 0.5)
		p95 = CalculatePercentile(latencias, 0.95)
		p99 = CalculatePercentile(latencias, 0.99)
	}

	errorRate := float64(countErrors) / float64(len(results)) * 100
	rps := float64(len(results)) / totalElapsed.Seconds()

	return Report{
		URL:           url,
		TotalRequests: len(results),
		RPS:           rps,
		ErrorRate:     errorRate,
		Status4xx5xx:  count4xx5xx,
		Timeouts:      countTimeouts,
		P50:           p50,
		P95:           p95,
		P99:           p99,
		Min:           min,
		Max:           max,
	}
}
