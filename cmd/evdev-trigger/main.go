package main

import (
	"context"
	"os"
	"sync"

	"github.com/hareku/evdev-trigger/pkg/config"
	"github.com/hareku/evdev-trigger/pkg/evdev"
	"github.com/hareku/evdev-trigger/pkg/notify"
	"github.com/hareku/evdev-trigger/pkg/watch"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
)

func main() {
	app := &cli.App{
		Name:  "evdev-trigger",
		Usage: "Trigger commands by evdev events.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "config",
				Usage:    "a configuration file path",
				Aliases:  []string{"conf", "c"},
				Required: true,
			},
			&cli.BoolFlag{
				Name:    "debug",
				Usage:   "debug mode flag",
				Aliases: []string{"d"},
			},
		},
		Action: func(c *cli.Context) error {
			ctx := c.Context
			logger := watch.NewLogger(os.Stdout, c.Bool("debug"))

			conf, err := config.Read(c.String("config"))
			if err != nil {
				logger.Errorf("Config error %q: %s", c.String("config"), err)
				return err
			}

			cnd := sync.NewCond(new(sync.Mutex))
			eg, ctx := errgroup.WithContext(ctx)
			eg.Go(func() error {
				return notify.NewFsNotifier().Subscribe(ctx, cnd)
			})
			eg.Go(func() error {
				return watch.NewWatcher(watch.NewWatcherInput{
					Phys:   conf.Phys,
					Logger: logger,
					Finder: evdev.NewFinder(),
					Handler: watch.NewHandler(watch.NewHandlerInput{
						Logger:   logger,
						Executor: watch.NewExecutor(),
						Triggers: conf.Triggers,
					}),
					ReconnectCond: cnd,
				}).Run(ctx)
			})

			return eg.Wait()
		},
	}

	err := app.RunContext(context.Background(), os.Args)
	if err != nil {
		os.Exit(1)
	}
}
