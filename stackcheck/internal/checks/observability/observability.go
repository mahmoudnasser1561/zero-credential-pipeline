package observability

import (
	"github.com/mahmoudnasser1561/zero-credential-pipeline/stackcheck/internal/check"
	"github.com/mahmoudnasser1561/zero-credential-pipeline/stackcheck/internal/executor"
)

// AllChecks returns every check for this stack, bound to a single target.
func AllChecks(exec executor.Executor, target string) []check.Checker {
	return []check.Checker{
		NewNodeExporterCheck(exec, target),
		NewPrometheusHealthCheck(exec, target),
		NewGrafanaProxyChainCheck(exec, target),
	}
}
