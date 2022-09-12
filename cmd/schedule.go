/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"

	"github.com/spf13/cobra"
)

// scheduleCmd represents the schedule command
var scheduleCmd = &cobra.Command{
	Use:   "schedule",
	Short: "Schedule an interval in minutes to run the task list",
	Long: `Schedule an interval in minutes to run the task list:
"dsync schedule [minutes] [-d|--del]"
If [-d|-del] flag is entered, [minutes] argument will be ignored.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		//create crontab command file
		home := os.Getenv("HOME")
		if err := os.Mkdir(path.Join(home, ".dsync"), 0750); err != nil && !os.IsExist(err) {
			log.Fatal(err)
		}
		if err := os.WriteFile(path.Join(home, ".dsync", "dsync.sh"), []byte(`#!/bin/zsh
. $HOME/.zshrc
dsync all
`), 0750); err != nil {
			log.Fatal(err)
		}
		//remove any dsync entry in crontab
		listCrontab := exec.Command("crontab", "-l")
		filterCrontab := exec.Command("grep", "-v", "dsync")
		cleanCrontab := exec.Command("crontab", "-")

		filterCrontab.Stdin, _ = listCrontab.StdoutPipe()
		cleanCrontab.Stdin, _ = filterCrontab.StdoutPipe()

		filterCrontab.Start()
		cleanCrontab.Start()
		listCrontab.Run()
		filterCrontab.Wait()
		cleanCrontab.Wait()

		if willDel, _ := cmd.Flags().GetBool("del"); willDel {
			return
		}
		//add a dsync entry in crontab
		if len(args) == 0 {
			log.Fatalln("Error, missing 'minutes' argument")
		}
		listOldCrontab := exec.Command("crontab", "-l")
		updateCrontab := exec.Command("crontab", "-")

		list, _ := listOldCrontab.Output()

		addNewCommand := exec.Command("echo", fmt.Sprintf("%v*/%v * * * * %v/.dsync/dsync.sh", string(list), args[0], home))

		updateCrontab.Stdin, _ = addNewCommand.StdoutPipe()

		updateCrontab.Start()
		addNewCommand.Start()
		listOldCrontab.Run()
		addNewCommand.Wait()
		updateCrontab.Wait()

	},
}

func init() {
	rootCmd.AddCommand(scheduleCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// scheduleCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	scheduleCmd.Flags().BoolP("del", "d", false, "Delete the schedule task run interval, if -d|--del flag is set arguments of shcedule command will be ignored")
}
