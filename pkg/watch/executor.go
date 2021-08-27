package watch

import (
	"context"
	"os/exec"

	"github.com/hareku/evdev-trigger/pkg/config"
)

//go:generate mockgen -source=${GOFILE} -destination=./${GOPACKAGE}mock/mock_${GOFILE} -package=${GOPACKAGE}mock

type Executor interface {
	Do(ctx context.Context, cmd config.Command) ([]byte, error)
}

type executor struct{}

func NewExecutor() Executor {
	return &executor{}
}

func (e *executor) Do(ctx context.Context, cmd config.Command) ([]byte, error) {
	var ecmd *exec.Cmd
	if len(cmd) == 1 {
		ecmd = exec.CommandContext(ctx, cmd[0])
	} else {
		ecmd = exec.CommandContext(ctx, cmd[0], cmd[1:]...)
	}

	return ecmd.Output()
}
