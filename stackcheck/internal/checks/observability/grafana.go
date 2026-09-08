package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mahmoudnasser1561/zero-credential-pipeline/stackcheck/internal/check"
	"github.com/mahmoudnasser1561/zero-credential-pipeline/stackcheck/internal/executor"
)

type grafanaProxyChainCheck struct {
	exec   executor.Executor
	target string
}

// NewGrafanaProxyChainCheck proves the observability chain end-to-end: it
// queries Grafana's own datasource-proxy endpoint (the same path a real
// dashboard panel uses), which forwards through to Prometheus, which
// answers from data it scraped from node_exporter. A pass here means the
// chain is genuinely wired together, not just that three ports each
// independently answer 200.
func NewGrafanaProxyChainCheck(exec executor.Executor, target string) check.Checker {
	return &grafanaProxyChainCheck{exec: exec, target: target}
}

func (c *grafanaProxyChainCheck) Name() string { return "grafana_datasource_chain" }

type grafanaDatasource struct {
	UID  string `json:"uid"`
	Type string `json:"type"`
}

type promQueryResponse struct {
	Status string `json:"status"`
	Data   struct {
		Result []struct {
			Value []any `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

func (c *grafanaProxyChainCheck) Check(ctx context.Context) check.Result {
	start := time.Now()

	dsOut, err := c.exec.Run(ctx, `curl -s -u admin:admin http://localhost:3000/api/datasources`)
	if err != nil {
		return check.Result{Name: c.Name(), Target: c.target, Status: check.Fail,
			Detail: fmt.Sprintf("executor error listing datasources: %v", err), Latency: time.Since(start)}
	}

	var datasources []grafanaDatasource
	if err := json.Unmarshal([]byte(dsOut.Stdout), &datasources); err != nil {
		return check.Result{Name: c.Name(), Target: c.target, Status: check.Fail,
			Detail: fmt.Sprintf("could not parse datasources response: %v", err), Latency: time.Since(start)}
	}

	var promUID string
	for _, ds := range datasources {
		if ds.Type == "prometheus" {
			promUID = ds.UID
			break
		}
	}
	if promUID == "" {
		return check.Result{Name: c.Name(), Target: c.target, Status: check.Fail,
			Detail: "no prometheus datasource found on grafana", Latency: time.Since(start)}
	}

	// -G --data-urlencode lets curl URL-encode the query itself, rather than
	// splicing `up{job="node"}` into the URL string by hand — the raw form
	// doesn't survive the shell -> curl -> HTTP quoting layers intact.
	cmd := fmt.Sprintf(
		`curl -s -u admin:admin -G --data-urlencode 'query=up{job="node"}' http://localhost:3000/api/datasources/proxy/uid/%s/api/v1/query`,
		promUID,
	)
	queryOut, err := c.exec.Run(ctx, cmd)
	if err != nil {
		return check.Result{Name: c.Name(), Target: c.target, Status: check.Fail,
			Detail: fmt.Sprintf("executor error querying proxy: %v", err), Latency: time.Since(start)}
	}

	var resp promQueryResponse
	if err := json.Unmarshal([]byte(queryOut.Stdout), &resp); err != nil {
		return check.Result{Name: c.Name(), Target: c.target, Status: check.Fail,
			Detail: fmt.Sprintf("could not parse datasource-proxy response: %v", err), Latency: time.Since(start)}
	}

	if resp.Status != "success" {
		return check.Result{Name: c.Name(), Target: c.target, Status: check.Fail,
			Detail: fmt.Sprintf("proxy query status was %q, not success", resp.Status), Latency: time.Since(start)}
	}

	if len(resp.Data.Result) == 0 {
		return check.Result{Name: c.Name(), Target: c.target, Status: check.Fail,
			Detail: "proxy query returned empty result set", Latency: time.Since(start)}
	}

	value := resp.Data.Result[0].Value
	if len(value) != 2 || fmt.Sprint(value[1]) != "1" {
		return check.Result{Name: c.Name(), Target: c.target, Status: check.Fail,
			Detail: fmt.Sprintf("expected up=1 for job=node via Grafana proxy, got %v", value), Latency: time.Since(start)}
	}

	return check.Result{Name: c.Name(), Target: c.target, Status: check.Pass,
		Detail: "Grafana -> Prometheus -> node_exporter chain returned live up=1", Latency: time.Since(start)}
}
