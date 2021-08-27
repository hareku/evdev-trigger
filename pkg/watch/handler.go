package watch

import (
	"context"
	"strings"
	"time"

	"github.com/hareku/evdev-trigger/pkg/config"
	"github.com/hareku/evdev-trigger/pkg/evdev"
)

//go:generate mockgen -source=${GOFILE} -destination=./${GOPACKAGE}mock/mock_${GOFILE} -package=${GOPACKAGE}mock

type Handler interface {
	Do(ctx context.Context, ev *evdev.InputEvent)
}

type NewHandlerInput struct {
	Logger   Logger
	Executor Executor
	Triggers map[uint16]config.CommandConfig
}

func NewHandler(in NewHandlerInput) Handler {
	return &handler{
		logger:   in.Logger,
		executor: in.Executor,
		triggers: in.Triggers,
		prev:     make(map[uint16]time.Time),
	}
}

type handler struct {
	logger   Logger
	executor Executor
	triggers map[uint16]config.CommandConfig
	prev     map[uint16]time.Time
}

func (h *handler) Do(ctx context.Context, ev *evdev.InputEvent) {
	h.logger.Debugf("Got input event: %v", ev)

	if ev.Type != evdev.EV_KEY {
		h.logger.Debugf("Event type is not EV_KEY(%d), got %d", evdev.EV_KEY, ev.Type)
		return
	}
	if ev.Value == 1 {
		h.logger.Debugf("Input key is still pressing (value is %d)", ev.Value)
		return
	}

	cmd, ok := h.triggers[ev.Code]
	if !ok {
		h.logger.Debugf("Trigger nof found for code(%d)", ev.Code)
		return
	}

	prev, ok := h.prev[ev.Code]
	if ok && time.Since(prev) < cmd.Interval {
		h.logger.Debugf("Skipped for interval, it will takes %v until the next run", cmd.Interval-time.Since(prev))
		return
	}
	h.prev[ev.Code] = time.Now()
	h.exec(ctx, cmd.Command)
}

func (h *handler) exec(ctx context.Context, cmd config.Command) {
	b, err := h.executor.Do(ctx, cmd)
	cmdStr := strings.Join(cmd, " ")
	if err != nil {
		h.logger.Errorf("Command %q failed: %s", cmdStr, err)
		return
	}
	if len(b) > 0 {
		h.logger.Infof("Command %q succeeded: %s", cmdStr, string(b))
		return
	}
	h.logger.Infof("Command %q succeeded, no output", cmdStr)
}
