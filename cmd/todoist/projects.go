package main

import (
	"os"

	"github.com/buddyh/todoist-cli/internal/api"
	"github.com/buddyh/todoist-cli/internal/output"
	"github.com/spf13/cobra"
)

func newProjectsCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "projects",
		Aliases: []string{"project", "proj"},
		Short:   "List all projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := output.NewFormatter(os.Stdout, flags.asJSON)

			client, err := getClient()
			if err != nil {
				return err
			}

			projects, err := client.GetProjects()
			if err != nil {
				return err
			}

			return out.WriteProjects(projects)
		},
	}

	// Add project add subcommand
	cmd.AddCommand(newProjectAddCmd(flags))

	return cmd
}

func newProjectAddCmd(flags *rootFlags) *cobra.Command {
	var (
		color    string
		favorite bool
	)

	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Create a new project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := output.NewFormatter(os.Stdout, flags.asJSON)

			client, err := getClient()
			if err != nil {
				return err
			}

			params := api.AddProjectParams{
				Name:       args[0],
				Color:      color,
				IsFavorite: favorite,
			}

			project, err := client.AddProject(params)
			if err != nil {
				return err
			}

			return out.WriteProject(project)
		},
	}

	cmd.Flags().StringVar(&color, "color", "", "project color")
	cmd.Flags().BoolVar(&favorite, "favorite", false, "mark as favorite")

	return cmd
}
