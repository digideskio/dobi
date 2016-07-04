package cmd

import (
	"fmt"
	"github.com/dnephin/dobi/config"
	"github.com/dnephin/dobi/tasks"
	"github.com/spf13/cobra"
)

func newCleanCommand(opts *dobiOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "autoclean",
		Short: "Run the remove action for all resources",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runClean(opts)
		},
	}
	return cmd
}

func runClean(opts *dobiOptions) error {
	conf, err := config.Load(opts.filename)
	if err != nil {
		return err
	}

	client, err := buildClient()
	if err != nil {
		return fmt.Errorf("Failed to create client: %s", err)
	}

	return tasks.Run(tasks.RunOptions{
		Client: client,
		Config: conf,
		Tasks:  removeTasks(opts.tasks),
		Quiet:  opts.quiet,
	})
}
