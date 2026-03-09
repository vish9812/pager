package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"pager/cache"
	"pager/pagerduty"
)

var pdClient *pagerduty.Client

var rootCmd = &cobra.Command{
	Use:   "pager",
	Short: "PagerDuty CLI tool",
	Long: `A CLI tool for managing PagerDuty schedules and overrides.

Requires a PagerDuty User API Token set as the PAGERDUTY_TOKEN environment variable.

To create a User API Token:
  1. Click your profile icon in the top-right corner of the PagerDuty web app.
  2. Select My Profile, then navigate to the User Settings tab.
  3. In the API Access section, click Create API User Token.
  4. Enter a descriptive name (e.g., "pager CLI") in the Description field.
  5. Click Create Token.
  6. Copy and securely store the token immediately — it will not be shown again.

Then export it before running pager:
  export PAGERDUTY_TOKEN=<your-token>`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip token check for cache commands
		if cmd.Name() == "clear" {
			return nil
		}
		token := os.Getenv("PAGERDUTY_TOKEN")
		if token == "" {
			return fmt.Errorf("PAGERDUTY_TOKEN environment variable is required")
		}
		pdClient = pagerduty.NewClient(token)
		return nil
	},
}

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Manage local cache of schedules and users",
}

var cachePathCmd = &cobra.Command{
	Use:   "path",
	Short: "Print the cache directory path",
	Long:  "Print the path to the local cache directory (~/.cache/pager).",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := cache.Dir()
		if err != nil {
			return err
		}
		fmt.Println(dir)
		return nil
	},
}

var cacheClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear the local cache",
	Long:  "Delete cached schedules and users data. The next command will fetch fresh data from PagerDuty.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := cache.Clear(); err != nil {
			return err
		}
		fmt.Println("Cache cleared.")
		return nil
	},
}

func init() {
	cacheCmd.AddCommand(cacheClearCmd)
	cacheCmd.AddCommand(cachePathCmd)
	rootCmd.AddCommand(cacheCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
