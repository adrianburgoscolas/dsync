/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a file|dir to the tasks list",
	Long: `Add a file or a directory to the sync tasks list:
"dsync add [file|dir]".`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fileToAdd, err := filepath.Abs(args[0])
		if err != nil {
			log.Fatalf("Unable to get file or directory %q: %v", args[0], err)
		}
		listFile, err := os.OpenFile(TasksFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("Unable to update/create tasks list file: %v", err)
		}
		defer listFile.Close()
		if _, err := fmt.Fprintf(listFile, "%v\n", fileToAdd); err != nil {
			log.Fatalf("Unable to write to tasks list file: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(addCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// addCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// addCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
