package check

import (
	"context"
	"time"
)

type Status string

const (
	Pass Status = "pass"
	Fail Status = "fail"
	Skip Status = "skip"
)

type Result struct {
	Name    string
	Target  string
	Status  Status
	Detail  string
	Latency time.Duration
}

type Checker interface {
	Name() string
	Check(ctx context.Context) Result
}
