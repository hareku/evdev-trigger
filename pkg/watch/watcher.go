package watch

import (
	"context"
	"errors"

	"sync"

	"github.com/hareku/evdev-trigger/pkg/evdev"
)

type Watcher interface {
	Run(ctx context.Context) error
}

type NewWatcherInput struct {
	Phys          string
	Logger        Logger
	Finder        evdev.Finder
	ReconnectCond *sync.Cond
	Handler       Handler
}

func NewWatcher(in NewWatcherInput) Watcher {
	return &watcher{
		phys:    in.Phys,
		logger:  in.Logger,
		finder:  in.Finder,
		handler: in.Handler,
		cnd:     in.ReconnectCond,
	}
}

type watcher struct {
	phys    string
	logger  Logger
	finder  evdev.Finder
	cnd     *sync.Cond
	handler Handler
	d       evdev.Device
}

func (w *watcher) Run(ctx context.Context) error {
	if err := w.run(ctx); err != nil {
		if errors.Is(err, context.Canceled) {
			w.logger.Infof("Terminated")
			return err
		}

		w.logger.Errorf("%+v", err)
		return err
	}

	return nil
}

func (w *watcher) run(ctx context.Context) error {
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
}

func (w *watcher) waitConnect(ctx context.Context) error {
	w.cnd.L.Lock()
	defer w.cnd.L.Unlock()

	ok, err := w.connect(ctx)
	for !ok {
		if err != nil && !errors.Is(err, evdev.ErrDeviceNotFound) {
			return err
		}
		w.logger.Debugf("Device not found (%s), waiting device connection.", w.phys)
		w.cnd.Wait()
		ok, err = w.connect(ctx)
	}

	w.logger.Debugf("Connected to %s", w.phys)
	return nil
}

func (w *watcher) connect(ctx context.Context) (bool, error) {
	w.logger.Debugf("Trying to connect %s", w.phys)

	device, err := w.finder.Find(w.phys)
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
			w.handler.Do(ctx, read.ev)
		}
	}
}
