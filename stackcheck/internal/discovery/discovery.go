package discovery

import "context"

type Discoverer interface {
	Discover(ctx context.Context) ([]string, error)
}
