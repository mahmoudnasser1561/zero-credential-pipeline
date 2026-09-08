package report

import (
	"fmt"
	"strings"
	"time"

	"github.com/mahmoudnasser1561/zero-credential-pipeline/stackcheck/internal/check"
)

const (
	ansiGreen = "\033[32m"
	ansiRed   = "\033[31m"
	ansiReset = "\033[0m"
)

type TableReporter struct{}

func NewTable() *TableReporter {
	return &TableReporter{}
}

func (t *TableReporter) Render(results []check.Result) string {
	var b strings.Builder

	b.WriteString("STACKCHECK REPORT\n")
	b.WriteString("==================\n")
	b.WriteString(fmt.Sprintf("%-28s %-6s %-9s %-9s %s\n", "CHECK", "TARGET", "STATUS", "LATENCY", "DETAIL"))

	passed := 0
	for _, r := range results {
		status := strings.ToUpper(string(r.Status))
		color := ansiGreen
		if r.Status == check.Fail {
			color = ansiRed
		} else if r.Status == check.Pass {
			passed++
		}

		b.WriteString(fmt.Sprintf("%-28s %-6s %s%-9s%s %-9s %s\n",
			r.Name, shortTarget(r.Target), color, status, ansiReset,
			r.Latency.Round(time.Millisecond).String(), r.Detail))
	}

	b.WriteString(fmt.Sprintf("\nSUMMARY: %d/%d checks passed\n", passed, len(results)))

	return b.String()
}

func shortTarget(target string) string {
	if len(target) > 6 {
		return target[len(target)-6:]
	}
	return target
}
