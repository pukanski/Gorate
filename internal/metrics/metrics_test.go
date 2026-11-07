package metrics

import (
	"errors"
	"testing"
	"time"
)

func TestCalculatePercentile(t *testing.T) {
	tests := []struct {
		name             string
		latenciaOrdenada []time.Duration
		p                float64
		wantIndex        time.Duration
	}{
		{
			name:             "Deve retornar 0 com o slice vazio",
			latenciaOrdenada: []time.Duration{},
			p:                0.99,
			wantIndex:        0 * time.Millisecond,
		},
		{
			name:             "Deve retornar 50ms com o p = 0.5",
			latenciaOrdenada: []time.Duration{10 * time.Millisecond, 20 * time.Millisecond, 30 * time.Millisecond, 40 * time.Millisecond, 50 * time.Millisecond, 60 * time.Millisecond, 70 * time.Millisecond, 80 * time.Millisecond, 90 * time.Millisecond, 100 * time.Millisecond},
			p:                0.5,
			wantIndex:        50 * time.Millisecond,
		},
		{
			name:             "Deve retornar 100ms com o p = 0.95",
			latenciaOrdenada: []time.Duration{10 * time.Millisecond, 20 * time.Millisecond, 30 * time.Millisecond, 40 * time.Millisecond, 50 * time.Millisecond, 60 * time.Millisecond, 70 * time.Millisecond, 80 * time.Millisecond, 90 * time.Millisecond, 100 * time.Millisecond},
			p:                0.95,
			wantIndex:        100 * time.Millisecond,
		},
		{
			name:             "Deve retornar 100ms com o p = 0.99",
			latenciaOrdenada: []time.Duration{10 * time.Millisecond, 20 * time.Millisecond, 30 * time.Millisecond, 40 * time.Millisecond, 50 * time.Millisecond, 60 * time.Millisecond, 70 * time.Millisecond, 80 * time.Millisecond, 90 * time.Millisecond, 100 * time.Millisecond},
			p:                0.99,
			wantIndex:        100 * time.Millisecond,
		},
		{
			name:             "Deve retornar 10ms com o p = 0.0 (retorna o primeiro elemento)",
			latenciaOrdenada: []time.Duration{10 * time.Millisecond, 20 * time.Millisecond, 30 * time.Millisecond, 40 * time.Millisecond, 50 * time.Millisecond, 60 * time.Millisecond, 70 * time.Millisecond, 80 * time.Millisecond, 90 * time.Millisecond, 100 * time.Millisecond},
			p:                0.0,
			wantIndex:        10 * time.Millisecond,
		},
		{
			name:             "Deve retornar 100ms com o p = 1.5 (retorna o último elemento)",
			latenciaOrdenada: []time.Duration{10 * time.Millisecond, 20 * time.Millisecond, 30 * time.Millisecond, 40 * time.Millisecond, 50 * time.Millisecond, 60 * time.Millisecond, 70 * time.Millisecond, 80 * time.Millisecond, 90 * time.Millisecond, 100 * time.Millisecond},
			p:                1.5,
			wantIndex:        100 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIndex := CalculatePercentile(tt.latenciaOrdenada, tt.p)
			if gotIndex != time.Duration(tt.wantIndex) {
				t.Errorf("CalculatePercentile() %s = %d, want = %d", tt.name, gotIndex, tt.wantIndex)
			}
		})
	}

}

func TestGenerateReport(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		results      []RequestResult
		totalElapsed time.Duration
		wantReport   Report
	}{
		{
			name:         "Deve retornar apenas o URL com o results vazio",
			url:          "https://www.google.com/",
			results:      []RequestResult{},
			totalElapsed: 20 * time.Millisecond,
			wantReport:   Report{URL: "https://www.google.com/"},
		},
		{
			name: "Deve retornar apenas o URL se o totalElapsed for igual ou menor que 0",
			url:  "https://www.google.com/",
			results: []RequestResult{
				{
					URL:        "https://www.google.com/",
					Duration:   0,
					StatusCode: 200,
					IsTimeout:  false,
					Err:        nil,
				},
			},
			totalElapsed: 0,
			wantReport:   Report{URL: "https://www.google.com/"},
		},
		{
			name: "Deve retornar corretamente o Report",
			url:  "https://www.google.com/",
			results: []RequestResult{
				{
					URL:        "https://www.google.com/",
					Duration:   100 * time.Millisecond,
					StatusCode: 200,
					IsTimeout:  false,
					Err:        nil,
				},
				{
					URL:        "https://g1.globo.com/",
					Duration:   300 * time.Millisecond,
					StatusCode: 200,
					IsTimeout:  false,
					Err:        nil,
				},
				{
					URL:        "https://www.youtube.com/",
					Duration:   200 * time.Millisecond,
					StatusCode: 500,
					IsTimeout:  false,
					Err:        nil,
				},
				{
					URL:        "https://www.quantamagazine.org/",
					Duration:   5000 * time.Millisecond,
					StatusCode: 200,
					IsTimeout:  true,
					Err:        errors.New("context deadline exceeded"),
				},
			},
			totalElapsed: 2 * time.Second,
			wantReport: Report{
				URL:           "https://www.google.com/",
				TotalRequests: 4,
				RPS:           2.0,
				ErrorRate:     50.0,
				Status4xx5xx:  1,
				Timeouts:      1,
				P50:           200 * time.Millisecond,
				P95:           300 * time.Millisecond,
				P99:           300 * time.Millisecond,
				Min:           100 * time.Millisecond,
				Max:           300 * time.Millisecond,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotReport := GenerateReport(tt.url, tt.results, tt.totalElapsed)

			if gotReport != tt.wantReport {
				t.Errorf("Generate Report() = %+v, want = %+v", gotReport, tt.wantReport)
			}
		})
	}

}
