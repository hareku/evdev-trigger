package main

import (
	"context"
	"log"
	"os"

	"github.com/hareku/evdev-trigger/pkg/config"
	"github.com/hareku/evdev-trigger/pkg/evdev"
	"github.com/hareku/evdev-trigger/pkg/notify"
	"github.com/hareku/evdev-trigger/pkg/watch"
	"github.com/urfave/cli/v2"
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
			conf, err := config.Read(c.String("config"))
			if err != nil {
				log.Printf("Config error %q: %s", c.String("config"), err)
				return err
			}

			w := watch.NewWatcher(watch.NewWatcherInput{
				Conf:     conf,
				Logger:   watch.NewStdLogger(c.Bool("debug")),
				Notifier: notify.NewFsNotifier(),
				Finder:   evdev.NewFinder(),
			})

			return w.Run(c.Context)
		},
	}

	err := app.RunContext(context.Background(), os.Args)
	if err != nil {
		os.Exit(1)
	}
}
