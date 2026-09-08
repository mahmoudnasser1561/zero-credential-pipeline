package report

import "github.com/mahmoudnasser1561/zero-credential-pipeline/stackcheck/internal/check"

type Reporter interface {
	Render(results []check.Result) string
}

// Overall returns "fail" if any result failed, "pass" otherwise. Skip is
// neutral and never causes a failure on its own.
func Overall(results []check.Result) check.Status {
	for _, r := range results {
		if r.Status == check.Fail {
			return check.Fail
		}
	}
	return check.Pass
}

// ExitCode derives a process exit code from the overall status.
func ExitCode(results []check.Result) int {
	if Overall(results) == check.Fail {
		return 1
	}
	return 0
}
