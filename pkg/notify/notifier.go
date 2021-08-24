package notify

import (
	"context"
	"sync"
)

type Notifier interface {
	// Subscribe subscribes notifications and broadcasts to cond.
	Subscribe(ctx context.Context, cond *sync.Cond) error
}
