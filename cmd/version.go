/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version:     %s\n", Version)
		fmt.Printf("Build Time:  %s\n", BuildTime)
		fmt.Printf("Git Commit:  %s\n", CommitHash)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
