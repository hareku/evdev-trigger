package watch

import (
	"context"
	"errors"
	"os/exec"
	"strings"

	"sync"
	"time"

	"github.com/hareku/evdev-trigger/pkg/config"
	"github.com/hareku/evdev-trigger/pkg/evdev"
	"github.com/hareku/evdev-trigger/pkg/notify"
	"golang.org/x/sync/errgroup"
)

type Watcher interface {
	Run(ctx context.Context) error
}

type NewWatcherInput struct {
	Conf     *config.Config
	Logger   Logger
	Notifier notify.Notifier
	Finder   evdev.Finder
}

func NewWatcher(in NewWatcherInput) Watcher {
	return &watcher{
		c:        in.Conf,
		logger:   in.Logger,
		notifier: in.Notifier,
		finder:   in.Finder,

		cnd: sync.NewCond(new(sync.Mutex)),
	}
}

type watcher struct {
	c      *config.Config
	logger Logger

	notifier notify.Notifier
	finder   evdev.Finder

	cnd  *sync.Cond
	d    evdev.Device
	prev time.Time
}

func (w *watcher) Run(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		return w.notifier.Subscribe(ctx, w.cnd)
	})

	eg.Go(func() error {
		for {
			err := w.listen(ctx)
			if err != nil {
				if errors.Is(err, errDeviceDisconnected) {
					if err := w.waitConnect(ctx); err != nil {
						return err
					}
					continue
				}
				return err
			}
		}
	})

	if err := eg.Wait(); err != nil {
		if errors.Is(err, context.Canceled) {
			w.logger.Infof("Terminated")
			return nil
		}
		w.logger.Errorf("%+v", err)
	}

	return nil
}

func (w *watcher) waitConnect(ctx context.Context) error {
	w.cnd.L.Lock()
	defer w.cnd.L.Unlock()

	ok, err := w.connect(ctx)
	for !ok {
		if err != nil && !errors.Is(err, evdev.ErrDeviceNotFound) {
			return err
		}
		w.logger.Debugf("Device not found (%s), waiting /dev/input events", w.c.Phys)
		w.cnd.Wait()
		ok, err = w.connect(ctx)
	}

	w.logger.Debugf("Connected to %s", w.c.Phys)
	return nil
}

func (w *watcher) connect(ctx context.Context) (bool, error) {
	w.logger.Debugf("Trying to connect %s", w.c.Phys)

	device, err := w.finder.Find(w.c.Phys)
	if err != nil {
		return false, err
	}
	w.d = device
	return true, nil
}

var errDeviceDisconnected = errors.New("device disconnected")

type read struct {
	ev  *evdev.InputEvent
	err error
}

func (w *watcher) listen(ctx context.Context) error {
	if w.d == nil {
		return errDeviceDisconnected
	}

	readCh := make(chan read)
	go func() {
		defer close(readCh)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				ev, err := w.d.Read()
				readCh <- read{
					ev:  ev,
					err: err,
				}
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case read, ok := <-readCh:
			if !ok {
				return nil
			}
			if read.err != nil {
				return errDeviceDisconnected
			}
			ev := read.ev

			if ev.Type != evdev.EV_KEY {
				w.logger.Debugf("Event type is not EV_KEY(%d), got %d", evdev.EV_KEY, ev.Type)
				continue
			}
			if ev.Value == 1 { // stil pressing
				continue
			}

			cmd, ok := w.c.Triggers[ev.Code]
			if !ok {
				w.logger.Debugf("Trigger nof found for code(%d)", ev.Code)
				continue
			}
			w.exec(ctx, cmd)
		}
	}
}

func (w *watcher) exec(ctx context.Context, cmd config.Command) {
	if time.Since(w.prev) < w.c.Interval {
		w.logger.Debugf("Skipped for interval, takes %v until the next run", w.c.Interval-time.Since(w.prev))
		return
	}
	w.prev = time.Now()

	var ecmd *exec.Cmd
	if len(cmd) == 1 {
		ecmd = exec.CommandContext(ctx, cmd[0])
	} else {
		ecmd = exec.CommandContext(ctx, cmd[0], cmd[1:]...)
	}

	b, err := ecmd.Output()
	cmdStr := strings.Join(cmd, " ")
	if err != nil {
		w.logger.Errorf("Command %q failed: %s", cmdStr, err)
		return
	}
	if len(b) > 0 {
		w.logger.Infof("Command %q succeeded: %s", cmdStr, string(b))
		return
	}
	w.logger.Infof("Command %q succeeded, no output", cmdStr)
}
