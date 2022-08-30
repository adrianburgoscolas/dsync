/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"github.com/spf13/cobra"
)

// authorizeCmd represents the authorize command
var authorizeCmd = &cobra.Command{
	Use:   "authorize",
	Short: "Get authorization to use user Google Drive",
	Long: `Get authorization to use user Google Drive:
"dsync authorize".`,
	Run: func(cmd *cobra.Command, args []string) {
		GetDriveService()
	},
}

func init() {
	rootCmd.AddCommand(authorizeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// authorizeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// authorizeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
