package main

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/hoahm-ts/awesome-ai-skills/pkg/config"
	"github.com/hoahm-ts/awesome-ai-skills/pkg/logger"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	log := logger.New(cfg.App.Name, cfg.App.Env, zerolog.InfoLevel)
	log.Info().Msg("starting worker")

	c, err := client.Dial(client.Options{
		HostPort:  cfg.Temporal.HostPort,
		Namespace: cfg.Temporal.Namespace,
	})
	if err != nil {
		return fmt.Errorf("dial temporal: %w", err)
	}
	defer c.Close()

	w := worker.New(c, cfg.Temporal.TaskQueue, worker.Options{})

	// Register workflows and activities here, for example:
	//   w.RegisterWorkflow(timeline.MyWorkflow)
	//   w.RegisterActivity(timeline.MyActivity)

	return w.Run(worker.InterruptCh())
}
