package observability

import (
	"context"

	"github.com/mahmoudnasser1561/zero-credential-pipeline/stackcheck/internal/executor"
)

// fakeExecutor lets check logic be tested without any real AWS/SSM call.
// Each call to Run pops the next canned output, in order.
type fakeExecutor struct {
	outputs []executor.Output
	err     error
	calls   int
}

func (f *fakeExecutor) Run(ctx context.Context, command string) (executor.Output, error) {
	if f.err != nil {
		return executor.Output{}, f.err
	}
	if f.calls >= len(f.outputs) {
		return executor.Output{}, nil
	}
	out := f.outputs[f.calls]
	f.calls++
	return out, nil
}
