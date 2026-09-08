package observability

import (
	"context"
	"testing"

	"github.com/mahmoudnasser1561/zero-credential-pipeline/stackcheck/internal/check"
	"github.com/mahmoudnasser1561/zero-credential-pipeline/stackcheck/internal/executor"
)

const healthyTargetsJSON = `{"status":"success","data":{"activeTargets":[
	{"labels":{"job":"prometheus"},"health":"up"},
	{"labels":{"job":"node"},"health":"up"}
]}}`

const oneDownTargetsJSON = `{"status":"success","data":{"activeTargets":[
	{"labels":{"job":"prometheus"},"health":"up"},
	{"labels":{"job":"node"},"health":"down"}
]}}`

func TestPrometheusHealthCheck(t *testing.T) {
	tests := []struct {
		name string
		exec *fakeExecutor
		want check.Status
	}{
		{
			name: "healthy endpoint and all targets up is a pass",
			exec: &fakeExecutor{outputs: []executor.Output{
				{Stdout: "200"},
				{Stdout: healthyTargetsJSON},
			}},
			want: check.Pass,
		},
		{
			name: "unhealthy endpoint is a fail",
			exec: &fakeExecutor{outputs: []executor.Output{
				{Stdout: "500"},
			}},
			want: check.Fail,
		},
		{
			name: "a down target is a fail",
			exec: &fakeExecutor{outputs: []executor.Output{
				{Stdout: "200"},
				{Stdout: oneDownTargetsJSON},
			}},
			want: check.Fail,
		},
		{
			name: "unparseable targets response is a fail",
			exec: &fakeExecutor{outputs: []executor.Output{
				{Stdout: "200"},
				{Stdout: "not json"},
			}},
			want: check.Fail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewPrometheusHealthCheck(tt.exec, "i-test")
			got := c.Check(context.Background())
			if got.Status != tt.want {
				t.Errorf("status = %v, want %v (detail: %s)", got.Status, tt.want, got.Detail)
			}
		})
	}
}
