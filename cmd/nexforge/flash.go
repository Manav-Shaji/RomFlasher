package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var flashCmd = &cobra.Command{
	Use:   "flash [partition] [image]",
	Short: "Flash an image to a specific partition",
	Long:  `Flash an Android partition using fastboot.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		partition := args[0]
		path := args[1]

		absPath, err := filepath.Abs(path)
		if err != nil {
			return err
		}

		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			return fmt.Errorf("image file not found: %s", absPath)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Flashing %s to %s...\n", absPath, partition)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		err = applicationContainer.Engine.FlashService.FlashImage(ctx, partition, absPath)
		if err != nil {
			return fmt.Errorf("flash failed: %w", err)
		}

		fmt.Fprintln(cmd.OutOrStdout(), "Flash complete!")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(flashCmd)
}
