package display

import (
	"fmt"
	"io"
	"text/tabwriter"
	"time"

	"github.com/pukanski/Gorate/internal/metrics"
)

func PrintReports(out io.Writer, reports []metrics.Report) {
	w := tabwriter.NewWriter(out, 0, 0, 3, ' ', 0)

	fmt.Fprintf(w, "URL\tRPS\tErros\t4xx/5xx\tTimeouts\tP50\tP95\tP99\tMin/Max\tTotal\n")

	for _, r := range reports {
		minMax := fmt.Sprintf("%v / %v", r.Min.Round(time.Millisecond), r.Max.Round(time.Millisecond))

		fmt.Fprintf(w, "%s\t%.2f\t%.1f%%\t%d\t%d\t%v\t%v\t%v\t%s\t%d\n",
			r.URL,
			r.RPS,
			r.ErrorRate,
			r.Status4xx5xx,
			r.Timeouts,
			r.P50.Round(time.Millisecond),
			r.P95.Round(time.Millisecond),
			r.P99.Round(time.Millisecond),
			minMax,
			r.TotalRequests,
		)
	}
	w.Flush()
}
