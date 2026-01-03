package main

import (
	"os"

	"github.com/buddyh/todoist-cli/internal/output"
	"github.com/spf13/cobra"
)

func newSectionsCmd(flags *rootFlags) *cobra.Command {
	var project string

	cmd := &cobra.Command{
		Use:     "sections",
		Aliases: []string{"section"},
		Short:   "List sections",
		Long: `List sections, optionally filtered by project.

Examples:
  todoist sections
  todoist sections -p Work`,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := output.NewFormatter(os.Stdout, flags.asJSON)

			client, err := getClient()
			if err != nil {
				return err
			}

			var projectID string
			if project != "" {
				p, err := client.FindProject(project)
				if err != nil {
					return err
				}
				projectID = p.ID
			}

			sections, err := client.GetSections(projectID)
			if err != nil {
				return err
			}

			return out.WriteSections(sections)
		},
	}

	cmd.Flags().StringVarP(&project, "project", "p", "", "filter by project name")

	// Add section add subcommand
	cmd.AddCommand(newSectionAddCmd(flags))

	return cmd
}

func newSectionAddCmd(flags *rootFlags) *cobra.Command {
	var project string

	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Create a new section in a project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := output.NewFormatter(os.Stdout, flags.asJSON)

			client, err := getClient()
			if err != nil {
				return err
			}

			p, err := client.FindProject(project)
			if err != nil {
				return err
			}

			section, err := client.AddSection(args[0], p.ID)
			if err != nil {
				return err
			}

			if flags.asJSON {
				return out.JSON(section)
			}
			out.WriteSuccess("Created section: " + section.Name)
			return nil
		},
	}

	cmd.Flags().StringVarP(&project, "project", "p", "", "project name (required)")
	cmd.MarkFlagRequired("project")

	return cmd
}
