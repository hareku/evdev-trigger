package notify

import (
	"context"
	"sync"
)

//go:generate mockgen -source=${GOFILE} -destination=./${GOPACKAGE}mock/mock_${GOFILE} -package=${GOPACKAGE}mock

type Notifier interface {
	// Subscribe subscribes notifications and broadcasts to cond.
	Subscribe(ctx context.Context, cond *sync.Cond) error
}
