package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var devicesCmd = &cobra.Command{
	Use:   "devices",
	Short: "List connected ADB and Fastboot devices",
	Long:  `Queries and prints currently connected devices using platform tools.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		fmt.Fprintln(cmd.OutOrStdout(), "Querying Fastboot devices...")
		if err := applicationContainer.Engine.RunCommand(ctx, "fastboot", "devices"); err != nil {
			applicationContainer.Logger.Warn("Fastboot devices check failed")
		}

		fmt.Fprintln(cmd.OutOrStdout(), "\nQuerying ADB devices...")
		if err := applicationContainer.Engine.RunCommand(ctx, "adb", "devices"); err != nil {
			applicationContainer.Logger.Warn("ADB devices check failed")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(devicesCmd)
}
