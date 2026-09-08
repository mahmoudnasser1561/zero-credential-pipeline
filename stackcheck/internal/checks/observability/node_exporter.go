package observability

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mahmoudnasser1561/zero-credential-pipeline/stackcheck/internal/check"
	"github.com/mahmoudnasser1561/zero-credential-pipeline/stackcheck/internal/executor"
)

type nodeExporterCheck struct {
	exec   executor.Executor
	target string
}

func NewNodeExporterCheck(exec executor.Executor, target string) check.Checker {
	return &nodeExporterCheck{exec: exec, target: target}
}

func (c *nodeExporterCheck) Name() string { return "node_exporter_health" }

func (c *nodeExporterCheck) Check(ctx context.Context) check.Result {
	start := time.Now()
	out, err := c.exec.Run(ctx, `curl -s -o /dev/null -w '%{http_code}' http://localhost:9100/metrics`)
	latency := time.Since(start)

	if err != nil {
		return check.Result{Name: c.Name(), Target: c.target, Status: check.Fail,
			Detail: fmt.Sprintf("executor error: %v", err), Latency: latency}
	}

	code := strings.TrimSpace(out.Stdout)
	if code != "200" {
		return check.Result{Name: c.Name(), Target: c.target, Status: check.Fail,
			Detail: fmt.Sprintf("expected http 200 from :9100/metrics, got %q", code), Latency: latency}
	}

	return check.Result{Name: c.Name(), Target: c.target, Status: check.Pass,
		Detail: "http 200 from :9100/metrics", Latency: latency}
}
