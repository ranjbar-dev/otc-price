package domain

import "context"

type LatestBarRepository interface {
	Save(ctx context.Context, bar Bar) error
}
