/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"log"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
)

//tasks list file used for all commands
var TasksFile = path.Join(UserHome, ".dsync/tasks.dsync")

//GetTasks returns a slice of all tasks to sync.
func GetTasks() []string {
	tasks, err := os.ReadFile(TasksFile)
	if err != nil {
		log.Fatalf("Unable to read file %q: %v", TasksFile, err)
	}
	return strings.Fields(string(tasks))
}

// allCmd represents the all command
var allCmd = &cobra.Command{
	Use:   "all",
	Short: "Run all sync tasks",
	Long: `Run all sync tasks added by the user:
"dsync all"
You can list all sync tasks by using:
"dsync list" command.`,
	Run: func(cmd *cobra.Command, args []string) {
		tasksSlice := GetTasks()
		for _, fileToSync := range tasksSlice {

			fileStats, err := os.Lstat(fileToSync)
			if err != nil {
				log.Fatalf("Unable to get file or dir %q stats: %v", args[0], err)
			}

			srv := GetDriveService()

			switch {
			case fileStats.Mode().IsDir():
				SyncDir(fileToSync, nil, srv)

			case fileStats.Mode().IsRegular():
				SyncFile(fileToSync, nil, srv)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(allCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// allCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// allCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
