package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version:     %s\n", Version)
		fmt.Printf("Build Time:  %s\n", BuildTime)
		fmt.Printf("Git Commit:  %s\n", CommitHash)
		fmt.Printf("Config File:  %s\n", viper.ConfigFileUsed())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
