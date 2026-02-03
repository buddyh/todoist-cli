package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/buddyh/todoist-cli/internal/api"
	"github.com/buddyh/todoist-cli/internal/config"
	"github.com/buddyh/todoist-cli/internal/output"
	"github.com/spf13/cobra"
)

func newAuthCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth [token]",
		Short: "Authenticate with Todoist API token",
		Long: `Authenticate with your Todoist API token.

Get your token from: https://todoist.com/app/settings/integrations/developer

You can either:
  1. Pass the token as an argument: todoist auth <token>
  2. Run interactively and paste when prompted: todoist auth
  3. Set TODOIST_API_TOKEN environment variable`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := output.NewFormatter(os.Stdout, flags.asJSON)
			var token string

			if len(args) > 0 {
				token = args[0]
			} else {
				// Interactive prompt
				fmt.Print("Enter your Todoist API token: ")
				reader := bufio.NewReader(os.Stdin)
				input, err := reader.ReadString('\n')
				if err != nil {
					return fmt.Errorf("failed to read token: %w", err)
				}
				token = strings.TrimSpace(input)
			}

			if token == "" {
				return fmt.Errorf("token cannot be empty")
			}

			// Validate token by making a test request
			client := api.NewClient(token)
			_, err := client.GetProjects()
			if err != nil {
				return fmt.Errorf("invalid token: %w", err)
			}

			// Save to config
			cfg := &config.Config{APIToken: token}
			if err := config.Save(cfg); err != nil {
				return err
			}

			configPath, _ := config.ConfigPath() // Error already handled by Save()
			out.WriteSuccess(fmt.Sprintf("Authenticated successfully. Config saved to %s", configPath))
			return nil
		},
	}

	// Add logout subcommand
	cmd.AddCommand(&cobra.Command{
		Use:   "logout",
		Short: "Remove stored credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := output.NewFormatter(os.Stdout, flags.asJSON)
			path, err := config.ConfigPath()
			if err != nil {
				return err
			}
			if err := os.Remove(path); err != nil {
				if os.IsNotExist(err) {
					out.WriteSuccess("No credentials stored.")
					return nil
				}
				return fmt.Errorf("failed to remove config: %w", err)
			}
			out.WriteSuccess("Logged out successfully.")
			return nil
		},
	})

	// Add status subcommand
	cmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Check authentication status",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := output.NewFormatter(os.Stdout, flags.asJSON)
			_, err := config.Load()
			if err != nil {
				out.WriteError(err)
				return nil
			}
			out.WriteSuccess("Authenticated")
			return nil
		},
	})

	return cmd
}
