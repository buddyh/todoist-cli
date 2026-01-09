package main

import (
	"os"

	"github.com/buddyh/todoist-cli/internal/api"
	"github.com/buddyh/todoist-cli/internal/config"
	"github.com/buddyh/todoist-cli/internal/output"
	"github.com/spf13/cobra"
)

var version = "dev"

type rootFlags struct {
	asJSON bool
}

func execute(args []string) error {
	var flags rootFlags

	rootCmd := &cobra.Command{
		Use:           "todoist",
		Short:         "Todoist CLI - manage tasks from the command line",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Default: show today's tasks
			return runTasks(cmd, &flags, true, "", "", false)
		},
	}
	rootCmd.SetVersionTemplate("todoist {{.Version}}\n")

	rootCmd.PersistentFlags().BoolVar(&flags.asJSON, "json", false, "output JSON instead of human-readable text")

	// Add subcommands
	rootCmd.AddCommand(newAuthCmd(&flags))
	rootCmd.AddCommand(newTasksCmd(&flags))
	rootCmd.AddCommand(newAddCmd(&flags))
	rootCmd.AddCommand(newCompleteCmd(&flags))
	rootCmd.AddCommand(newDoneCmd(&flags)) // alias for complete
	rootCmd.AddCommand(newDeleteCmd(&flags))
	rootCmd.AddCommand(newUpdateCmd(&flags))
	rootCmd.AddCommand(newProjectsCmd(&flags))
	rootCmd.AddCommand(newLabelsCmd(&flags))
	rootCmd.AddCommand(newSectionsCmd(&flags))
	rootCmd.AddCommand(newSearchCmd(&flags))
	rootCmd.AddCommand(newViewCmd(&flags))
	rootCmd.AddCommand(newCompletedCmd(&flags))
	rootCmd.AddCommand(newReopenCmd(&flags))
	rootCmd.AddCommand(newCommentCmd(&flags))
	rootCmd.AddCommand(newMoveCmd(&flags))

	rootCmd.SetArgs(args)
	if err := rootCmd.Execute(); err != nil {
		out := output.NewFormatter(os.Stderr, flags.asJSON)
		out.WriteError(err)
		return err
	}
	return nil
}

// getClient returns an authenticated API client
func getClient() (*api.Client, error) {
	token, err := config.GetToken()
	if err != nil {
		return nil, err
	}
	return api.NewClient(token), nil
}
