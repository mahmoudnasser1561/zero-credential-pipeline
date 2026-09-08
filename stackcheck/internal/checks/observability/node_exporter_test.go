package observability

import (
	"context"
	"errors"
	"testing"

	"github.com/mahmoudnasser1561/zero-credential-pipeline/stackcheck/internal/check"
	"github.com/mahmoudnasser1561/zero-credential-pipeline/stackcheck/internal/executor"
)

func TestNodeExporterCheck(t *testing.T) {
	tests := []struct {
		name string
		exec *fakeExecutor
		want check.Status
	}{
		{
			name: "200 is a pass",
			exec: &fakeExecutor{outputs: []executor.Output{{Stdout: "200"}}},
			want: check.Pass,
		},
		{
			name: "non-200 is a fail",
			exec: &fakeExecutor{outputs: []executor.Output{{Stdout: "503"}}},
			want: check.Fail,
		},
		{
			name: "executor error is a fail",
			exec: &fakeExecutor{err: errors.New("ssm timeout")},
			want: check.Fail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewNodeExporterCheck(tt.exec, "i-test")
			got := c.Check(context.Background())
			if got.Status != tt.want {
				t.Errorf("status = %v, want %v (detail: %s)", got.Status, tt.want, got.Detail)
			}
			if got.Target != "i-test" {
				t.Errorf("target = %v, want i-test", got.Target)
			}
		})
	}
}
