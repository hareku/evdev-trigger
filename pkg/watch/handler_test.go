package watch_test

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/hareku/evdev-trigger/pkg/config"
	"github.com/hareku/evdev-trigger/pkg/evdev"
	"github.com/hareku/evdev-trigger/pkg/watch"
	"github.com/hareku/evdev-trigger/pkg/watch/watchmock"
)

func Test_handler_Do(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	executor := watchmock.NewMockExecutor(ctrl)

	executor.EXPECT().Do(ctx, config.Command{"echo", "Hello", "World"}).Times(1)

	handler := watch.NewHandler(watch.NewHandlerInput{
		Logger:   watch.NewLogger(io.Discard, true),
		Executor: executor,
		Triggers: map[uint16]config.CommandConfig{
			10: {
				Command: config.Command{"echo", "Hello", "World"},
			},
		},
	})

	handler.Do(ctx, &evdev.InputEvent{
		Type:  evdev.EV_KEY,
		Code:  10,
		Value: 1,
	})
	handler.Do(ctx, &evdev.InputEvent{
		Type:  evdev.EV_KEY,
		Code:  10,
		Value: 0,
	})

	handler.Do(ctx, &evdev.InputEvent{
		Type:  evdev.EV_SYN,
		Code:  10,
		Value: 0,
	})
}

func Test_handler_Do_WithInterval(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	executor := watchmock.NewMockExecutor(ctrl)

	executor.EXPECT().Do(ctx, config.Command{"echo", "Hello", "World"}).Times(2)

	handler := watch.NewHandler(watch.NewHandlerInput{
		Logger:   watch.NewLogger(io.Discard, true),
		Executor: executor,
		Triggers: map[uint16]config.CommandConfig{
			10: {
				Command:  config.Command{"echo", "Hello", "World"},
				Interval: time.Millisecond * 100,
			},
		},
	})

	handler.Do(ctx, &evdev.InputEvent{
		Type:  evdev.EV_KEY,
		Code:  10,
		Value: 0,
	})
	handler.Do(ctx, &evdev.InputEvent{
		Type:  evdev.EV_KEY,
		Code:  10,
		Value: 0,
	})
	time.Sleep(time.Millisecond * 100)
	handler.Do(ctx, &evdev.InputEvent{
		Type:  evdev.EV_KEY,
		Code:  10,
		Value: 0,
	})
}
