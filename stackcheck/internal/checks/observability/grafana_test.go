package observability

import (
	"context"
	"testing"

	"github.com/mahmoudnasser1561/zero-credential-pipeline/stackcheck/internal/check"
	"github.com/mahmoudnasser1561/zero-credential-pipeline/stackcheck/internal/executor"
)

const datasourcesListJSON = `[{"id":1,"uid":"PBFA97CFB590B2093","type":"prometheus","name":"Prometheus"}]`
const noDatasourcesJSON = `[]`

const upQueryResponseJSON = `{"status":"success","data":{"resultType":"vector","result":[
	{"metric":{"job":"node"},"value":[1700000000,"1"]}
]}}`

const downQueryResponseJSON = `{"status":"success","data":{"resultType":"vector","result":[
	{"metric":{"job":"node"},"value":[1700000000,"0"]}
]}}`

const emptyQueryResponseJSON = `{"status":"success","data":{"resultType":"vector","result":[]}}`

func TestGrafanaProxyChainCheck(t *testing.T) {
	tests := []struct {
		name string
		exec *fakeExecutor
		want check.Status
	}{
		{
			name: "live up=1 through the proxy is a pass",
			exec: &fakeExecutor{outputs: []executor.Output{
				{Stdout: datasourcesListJSON},
				{Stdout: upQueryResponseJSON},
			}},
			want: check.Pass,
		},
		{
			name: "up=0 through the proxy is a fail",
			exec: &fakeExecutor{outputs: []executor.Output{
				{Stdout: datasourcesListJSON},
				{Stdout: downQueryResponseJSON},
			}},
			want: check.Fail,
		},
		{
			name: "empty result set is a fail",
			exec: &fakeExecutor{outputs: []executor.Output{
				{Stdout: datasourcesListJSON},
				{Stdout: emptyQueryResponseJSON},
			}},
			want: check.Fail,
		},
		{
			name: "no prometheus datasource is a fail",
			exec: &fakeExecutor{outputs: []executor.Output{
				{Stdout: noDatasourcesJSON},
			}},
			want: check.Fail,
		},
		{
			name: "unparseable response is a fail",
			exec: &fakeExecutor{outputs: []executor.Output{
				{Stdout: datasourcesListJSON},
				{Stdout: "not json"},
			}},
			want: check.Fail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewGrafanaProxyChainCheck(tt.exec, "i-test")
			got := c.Check(context.Background())
			if got.Status != tt.want {
				t.Errorf("status = %v, want %v (detail: %s)", got.Status, tt.want, got.Detail)
			}
		})
	}
}
