package main

import (
	"os"

	"github.com/buddyh/todoist-cli/internal/output"
	"github.com/spf13/cobra"
)

func newLabelsCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "labels",
		Aliases: []string{"label", "tags"},
		Short:   "List all labels",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := output.NewFormatter(os.Stdout, flags.asJSON)

			client, err := getClient()
			if err != nil {
				return err
			}

			labels, err := client.GetLabels()
			if err != nil {
				return err
			}

			return out.WriteLabels(labels)
		},
	}

	// Add label add subcommand
	cmd.AddCommand(newLabelAddCmd(flags))

	return cmd
}

func newLabelAddCmd(flags *rootFlags) *cobra.Command {
	var color string

	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Create a new label",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := output.NewFormatter(os.Stdout, flags.asJSON)

			client, err := getClient()
			if err != nil {
				return err
			}

			label, err := client.AddLabel(args[0], color)
			if err != nil {
				return err
			}

			if flags.asJSON {
				return out.JSON(label)
			}
			out.WriteSuccess("Created label: @" + label.Name)
			return nil
		},
	}

	cmd.Flags().StringVar(&color, "color", "", "label color")

	return cmd
}
