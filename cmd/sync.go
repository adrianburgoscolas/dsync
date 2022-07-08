/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
)

var userHome, _ = os.UserHomeDir()

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := filepath.Join(userHome, ".dsync/token.json")
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

//Sync a folder recurrently to google drive.
func SyncDir(dir string, parent []string, srv *drive.Service) {

	//read current dir
	currentDirFiles, err := os.ReadDir(dir)
	if err != nil {
		log.Fatalf("Unable to read dir: %v", err)
	}

	driveFolderName := filepath.Base(dir)

	folderMeta := &drive.File{
		Name:     driveFolderName,
		MimeType: "application/vnd.google-apps.folder",
		Parents:  parent,
	}
	driveFolder, err := srv.Files.Create(folderMeta).Do()
	if err != nil {
		log.Fatalf("Unable to create Drive folder: %v", err)
	}

	for _, file := range currentDirFiles {
		if file.IsDir() {
			SyncDir(path.Join(dir, file.Name()), []string{driveFolder.Id}, srv)
		} else {
			fileMeta := &drive.File{
				Name:    file.Name(),
				Parents: []string{driveFolder.Id},
			}

			f, err := os.Open(path.Join(dir, file.Name()))
			if err != nil {
				log.Fatalf("Unable to open file %q %v", file.Name(), err)
			}

			defer f.Close()

			driveFile, err := srv.Files.Create(fileMeta).Media(f).Do()
			if err != nil {
				log.Fatalf("Unable to create %q in google drive %v", file.Name(), err)
			}

			fmt.Printf("Uploaded file %q Id %v to google drive\n", file.Name(), driveFile.Id)

		}
	}
}

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync a directory or a file",
	Long: `Sync/backup the current directory if no argument is passed
or sync/backup the file or directory specified in the args:
 - dsync sync [file/directory]`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		b, err := ioutil.ReadFile(filepath.Join(userHome, ".dsync/client_secret_654016737032-1jj92r0pcflivhq85nh31fim8fhlr1o7.apps.googleusercontent.com.json"))
		if err != nil {
			log.Fatalf("Unable to read client secret file: %v", err)
		}

		// If modifying these scopes, delete your previously saved token.json.
		config, err := google.ConfigFromJSON(b, drive.DriveFileScope)
		if err != nil {
			log.Fatalf("Unable to parse client secret file to config: %v", err)
		}
		client := getClient(config)

		srv, err := drive.New(client)
		if err != nil {
			log.Fatalf("Unable to retrieve Drive client: %v", err)
		}

		var dirToSync string
		if len(args) == 0 {
			currentDir, err := os.Getwd()
			if err != nil {
				log.Fatalf("Unable to get working directory: %v", err)
			}
			dirToSync = currentDir
		} else {
			dirToSync = args[0]
		}

		SyncDir(dirToSync, nil, srv)

	},
}

func init() {
	rootCmd.AddCommand(syncCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// syncCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// syncCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
