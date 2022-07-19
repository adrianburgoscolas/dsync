/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a file|dir from the tasks list",
	Long: `Remove a file or a directory from the sync tasks list:
"dsync remove [file|dir]".`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fileToRemove, err := filepath.Abs(args[0])
		if err != nil {
			log.Fatalf("Unable to get file or directory %q: %v", args[0], err)
		}
		listFileData, err := os.ReadFile(path.Join(UserHome, ".dsync/tasks.dsync"))
		if err != nil {
			log.Fatalf("Unable to read tasks list file: %v", err)
		}
		newList := strings.Replace(string(listFileData), fmt.Sprintf("%s\n", fileToRemove), "", -1)
		if err := os.WriteFile(path.Join(UserHome, ".dsync/tasks.dsync"), []byte(newList), 0644); err != nil {
			log.Fatalf("Unable to update tasks list file: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// removeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// removeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
