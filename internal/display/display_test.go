package display

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/pukanski/Gorate/internal/metrics"
)

func TestPrintReports(t *testing.T) {
	reports := []metrics.Report{
		{
			URL:           "https://www.google.com/",
			TotalRequests: 1500,
			RPS:           150.456,
			ErrorRate:     5.2,
			Status4xx5xx:  12,
			Timeouts:      3,
			P50:           120 * time.Millisecond,
			P95:           250 * time.Millisecond,
			P99:           500 * time.Millisecond,
			Min:           80 * time.Millisecond,
			Max:           800 * time.Millisecond,
		},
	}

	var buf bytes.Buffer

	PrintReports(&buf, reports)
	output := buf.String()

	if !strings.Contains(output, "URL") || !strings.Contains(output, "RPS") {
		t.Errorf("cabeçalho esperado não encontrado na saída:\n%s", output)
	}

	if !strings.Contains(output, "https://www.google.com/") {
		t.Errorf("a URL esperada não foi encontrada na saída:\n%s", output)
	}

	if !strings.Contains(output, "150.46") {
		t.Errorf("o RPS não foi formatado corretamente (esperado '150.46'):\n%s", output)
	}
	if !strings.Contains(output, "5.2%") {
		t.Errorf("o ErrorRate não foi formatado corretamente (esperado '5.2%%'):\n%s", output)
	}

	if !strings.Contains(output, "80ms / 800ms") {
		t.Errorf("coluna Min/Max não foi formatada corretamente:\n%s", output)
	}
}
