package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mahmoudnasser1561/zero-credential-pipeline/stackcheck/internal/check"
	"github.com/mahmoudnasser1561/zero-credential-pipeline/stackcheck/internal/executor"
)

type prometheusCheck struct {
	exec   executor.Executor
	target string
}

func NewPrometheusHealthCheck(exec executor.Executor, target string) check.Checker {
	return &prometheusCheck{exec: exec, target: target}
}

func (c *prometheusCheck) Name() string { return "prometheus_health" }

type promTargetsResponse struct {
	Data struct {
		ActiveTargets []struct {
			Labels map[string]string `json:"labels"`
			Health string            `json:"health"`
		} `json:"activeTargets"`
	} `json:"data"`
}

func (c *prometheusCheck) Check(ctx context.Context) check.Result {
	start := time.Now()

	healthOut, err := c.exec.Run(ctx, `curl -s -o /dev/null -w '%{http_code}' http://localhost:9090/-/healthy`)
	if err != nil {
		return check.Result{Name: c.Name(), Target: c.target, Status: check.Fail,
			Detail: fmt.Sprintf("executor error: %v", err), Latency: time.Since(start)}
	}
	if code := strings.TrimSpace(healthOut.Stdout); code != "200" {
		return check.Result{Name: c.Name(), Target: c.target, Status: check.Fail,
			Detail: fmt.Sprintf("expected http 200 from /-/healthy, got %q", code), Latency: time.Since(start)}
	}

	targetsOut, err := c.exec.Run(ctx, `curl -s http://localhost:9090/api/v1/targets`)
	if err != nil {
		return check.Result{Name: c.Name(), Target: c.target, Status: check.Fail,
			Detail: fmt.Sprintf("executor error fetching targets: %v", err), Latency: time.Since(start)}
	}

	var resp promTargetsResponse
	if err := json.Unmarshal([]byte(targetsOut.Stdout), &resp); err != nil {
		return check.Result{Name: c.Name(), Target: c.target, Status: check.Fail,
			Detail: fmt.Sprintf("could not parse /api/v1/targets response: %v", err), Latency: time.Since(start)}
	}

	if len(resp.Data.ActiveTargets) == 0 {
		return check.Result{Name: c.Name(), Target: c.target, Status: check.Fail,
			Detail: "no active scrape targets found", Latency: time.Since(start)}
	}

	for _, t := range resp.Data.ActiveTargets {
		if t.Health != "up" {
			return check.Result{Name: c.Name(), Target: c.target, Status: check.Fail,
				Detail: fmt.Sprintf("target job=%s is %s, not up", t.Labels["job"], t.Health), Latency: time.Since(start)}
		}
	}

	return check.Result{Name: c.Name(), Target: c.target, Status: check.Pass,
		Detail: fmt.Sprintf("%d scrape target(s) healthy", len(resp.Data.ActiveTargets)), Latency: time.Since(start)}
}
