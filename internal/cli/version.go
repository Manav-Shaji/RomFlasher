package cli

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of NexForge",
	Long:  `All software has versions. This is NexForge's.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(cmd.OutOrStdout(), "NexForge %s\n", Version)
		fmt.Fprintf(cmd.OutOrStdout(), "Commit: %s\n", Commit)
		fmt.Fprintf(cmd.OutOrStdout(), "Build Date: %s\n", BuildDate)
		fmt.Fprintf(cmd.OutOrStdout(), "Go Version: %s %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
