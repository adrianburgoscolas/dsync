/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all sync tasks",
	Long: `List all sync tasks:
"dsync list".`,
	Run: func(cmd *cobra.Command, args []string) {
		listFileData, err := os.ReadFile(TasksFile)
		if err != nil {
			log.Fatalf("Unable to read tasks list file: %v", err)
		}
		fmt.Println("Tasks List:")
		fmt.Println(string(listFileData))
		listCrontab := exec.Command("crontab", "-l")
		filterCrontab := exec.Command("grep", "dsync")

		filterCrontab.Stdin, _ = listCrontab.StdoutPipe()

		listCrontab.Start()
		filterOut, _ := filterCrontab.Output()
		listCrontab.Wait()

		if string(filterOut) == "" {
			fmt.Println("No task scheduled")
			return
		}
		task := strings.Fields(string(filterOut))[0][2:]
		fmt.Printf("Tasks runs every %v minutes.\n", task)

	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
