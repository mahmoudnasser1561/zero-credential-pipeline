package executor

import "context"

type Output struct {
	Stdout   string
	Stderr   string
	ExitCode int32
}

type Executor interface {
	Run(ctx context.Context, command string) (Output, error)
}
