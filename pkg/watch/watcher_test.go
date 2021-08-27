package watch_test

import (
	"context"
	"errors"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/hareku/evdev-trigger/pkg/evdev"
	"github.com/hareku/evdev-trigger/pkg/evdev/evdevmock"
	"github.com/hareku/evdev-trigger/pkg/watch"
	"github.com/hareku/evdev-trigger/pkg/watch/watchmock"
	"github.com/stretchr/testify/require"
)

func Test_watcher_Run_HandleInputEvent(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	phys := "00-00-00-00-00"

	handler := watchmock.NewMockHandler(ctrl)
	handler.EXPECT().Do(gomock.Any(), &evdev.InputEvent{
		Type:  evdev.EV_KEY,
		Code:  10,
		Value: 0,
	}).Times(1)

	device := evdevmock.NewMockDevice(ctrl)
	device.EXPECT().Read().Times(1).Return(&evdev.InputEvent{
		Type:  evdev.EV_KEY,
		Code:  10,
		Value: 0,
	}, nil)
	device.EXPECT().Read().Times(1).DoAndReturn(func() (*evdev.InputEvent, error) {
		<-ctx.Done()
		return nil, ctx.Err()
	})

	finder := evdevmock.NewMockFinder(ctrl)
	finder.EXPECT().Find(phys).Times(1).Return(device, nil)

	watcher := watch.NewWatcher(watch.NewWatcherInput{
		Phys:          phys,
		Logger:        watch.NewLogger(io.Discard, true),
		Finder:        finder,
		Handler:       handler,
		ReconnectCond: sync.NewCond(new(sync.Mutex)),
	})

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	err := watcher.Run(ctx)
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

func Test_watcher_Run_ReconnectDevice(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	phys := "00-00-00-00-00"
	cnd := sync.NewCond(new(sync.Mutex))

	device1 := evdevmock.NewMockDevice(ctrl)

	device1.EXPECT().Read().Times(2).DoAndReturn(func() (*evdev.InputEvent, error) {
		go func() {
			time.Sleep(time.Millisecond * 200)
			cnd.Broadcast()
		}()
		return nil, errors.New("disconnected")
	})
	device2 := evdevmock.NewMockDevice(ctrl)
	device2.EXPECT().Read().Times(1).DoAndReturn(func() (*evdev.InputEvent, error) {
		<-ctx.Done()
		return nil, ctx.Err()
	})

	finder := evdevmock.NewMockFinder(ctrl)
	gomock.InOrder(
		finder.EXPECT().Find(phys).Times(1).Return(device1, nil),
		finder.EXPECT().Find(phys).Times(1).Return(device2, nil),
	)

	watcher := watch.NewWatcher(watch.NewWatcherInput{
		Phys:          phys,
		Logger:        watch.NewLogger(io.Discard, true),
		Finder:        finder,
		Handler:       watchmock.NewMockHandler(ctrl),
		ReconnectCond: cnd,
	})

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	err := watcher.Run(ctx)
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

func Test_watcher_Run_ReconnectDeviceWithNotFoundOnce(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	phys := "00-00-00-00-00"
	cnd := sync.NewCond(new(sync.Mutex))

	device1 := evdevmock.NewMockDevice(ctrl)

	device1.EXPECT().Read().Times(2).DoAndReturn(func() (*evdev.InputEvent, error) {
		go func() {
			time.Sleep(time.Millisecond * 200)
			cnd.Broadcast()
			time.Sleep(time.Millisecond * 200)
			cnd.Broadcast()
		}()
		return nil, errors.New("disconnected")
	})
	device2 := evdevmock.NewMockDevice(ctrl)
	device2.EXPECT().Read().Times(1).DoAndReturn(func() (*evdev.InputEvent, error) {
		<-ctx.Done()
		return nil, ctx.Err()
	})

	finder := evdevmock.NewMockFinder(ctrl)
	gomock.InOrder(
		finder.EXPECT().Find(phys).Times(1).Return(device1, nil),
		finder.EXPECT().Find(phys).Times(1).Return(nil, evdev.ErrDeviceNotFound),
		finder.EXPECT().Find(phys).Times(1).Return(device2, nil),
	)

	watcher := watch.NewWatcher(watch.NewWatcherInput{
		Phys:          phys,
		Logger:        watch.NewLogger(io.Discard, true),
		Finder:        finder,
		Handler:       watchmock.NewMockHandler(ctrl),
		ReconnectCond: cnd,
	})

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	err := watcher.Run(ctx)
	require.ErrorIs(t, err, context.DeadlineExceeded)
}
