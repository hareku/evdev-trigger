package notify

import (
	"context"
	"fmt"
	"sync"

	"github.com/fsnotify/fsnotify"
)

func NewFsNotifier() Notifier {
	return &fsNotifier{}
}

type fsNotifier struct{}

func (n *fsNotifier) Subscribe(ctx context.Context, cnd *sync.Cond) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	if err := watcher.Add("/dev/input/"); err != nil {
		return fmt.Errorf("failed to watch dir /dev/input/: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if event.Op&fsnotify.Create == fsnotify.Create {
				cnd.Broadcast()
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			return fmt.Errorf("error in watcher: %w", err)
		}
	}
}
