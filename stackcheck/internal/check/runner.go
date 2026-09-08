package check

import (
	"context"
	"sync"
)

type Runner struct{}

func NewRunner() *Runner {
	return &Runner{}
}

func (r *Runner) Run(ctx context.Context, checkers []Checker) []Result {
	results := make([]Result, len(checkers))
	var wg sync.WaitGroup

	for i, c := range checkers {
		wg.Add(1)
		go func(i int, c Checker) {
			defer wg.Done()
			results[i] = c.Check(ctx)
		}(i, c)
	}

	wg.Wait()
	return results
}
