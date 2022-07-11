/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

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

//SyncDir sync/backup a folder recurrently to google drive.
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
			SyncFile(path.Join(dir, file.Name()), []string{driveFolder.Id}, srv)
		}
	}
}

//SyncFile sync/backup a file to Google Drive.
func SyncFile(file string, parent []string, srv *drive.Service) {
	if ChkSumFile(file) {
		fmt.Printf("File %q is backed up and hasn't been modified\n", file)
		return
	}

	f, err := os.Open(file)
	if err != nil {
		log.Fatalf("Unable to open file %q: %v", file, err)
	}
	defer f.Close()

	fileName := filepath.Base(file)
	fileDir := path.Dir(file)
	chkSumFile, err := os.Open(path.Join(fileDir + "/." + fileName + ".sha256sum"))

	if errors.Is(err, os.ErrNotExist) {
		fileMeta := &drive.File{
			Name:    fileName,
			Parents: parent,
		}
		driveFile, err := srv.Files.Create(fileMeta).Media(f).Do()
		if err != nil {
			log.Fatalf("Unable to create file %q in Google Drive: %v", fileName, err)
		}
		fmt.Printf("Uploaded file %q Id %v to Google Drive\n", file, driveFile.Id)
		CreateChkSum(file, driveFile.Id)
		return
	}

	if _, err := chkSumFile.Seek(65, 0); err != nil {
		log.Fatalf("Unable to get drive file Id: %v\n", err)
	}
	byteSlc := make([]byte, 16)
	var driveFileId []byte
	for {
		n, err := chkSumFile.Read(byteSlc)
		if n != 0 {
			driveFileId = append(driveFileId, byteSlc[:n]...)
		}
		if err != nil {
			break
		}
	}
	driveFile, err := srv.Files.Update(string(driveFileId), &drive.File{}).Media(f).Do()
	if err != nil {
		log.Fatalf("Unable to update file %q in Google Drive: %v", fileName, err)
	}

	fmt.Printf("Updated file %q Id %v in Google Drive\n", file, driveFile.Id)
	CreateChkSum(file, driveFile.Id)
}

//ChkSumFile check if the given file hasn't been modified or backed up.
func ChkSumFile(file string) bool {
	fileDir := path.Dir(file)
	fileName := filepath.Base(file)
	chkSumFile := path.Join(fileDir, "/."+fileName+".sha256sum")
	chkSumFileHandle, err := os.Open(chkSumFile)
	defer chkSumFileHandle.Close()
	if errors.Is(err, os.ErrNotExist) {
		return false
	}

	checksumHash := make([]byte, 64)
	n, err := chkSumFileHandle.Read(checksumHash)
	if err != nil {
		log.Fatalf("Unable to read checksum file hash: %v\n", err)
	}
	fmt.Printf("Readed %v bytes from file %v\n", n, chkSumFile)
	fileData, err := os.ReadFile(file)
	if err != nil {
		log.Fatalf("Unable to read file %q: %v\n", file, err)
	}
	fileHash := sha256.New()
	hashBytes, err := fileHash.Write(fileData)
	if err != nil {
		log.Fatalf("Unable to write data: %v\n", err)
	}
	fmt.Printf("Hashing %v bytes from %v\n", hashBytes, file)
	hashSlc := fmt.Sprintf("%x", fileHash.Sum(nil))
	if string(checksumHash) == hashSlc {
		return true
	}
	return false
}

//CreateChkSum create checksum file from given file.
func CreateChkSum(file, driveFileId string) {
	fileDir := path.Dir(file)
	fileName := filepath.Base(file)
	chkSumFileHandle, err := os.Create(fileDir + "/." + fileName + ".sha256sum")
	if err != nil {
		log.Fatalf("Unable to create sha256sum file: %v\n", err)
	}
	defer chkSumFileHandle.Close()
	fileData, err := os.ReadFile(file)
	if err != nil {
		log.Fatalf("Unable to read file %q: %v\n", file, err)
	}
	fileHash := sha256.New()
	if _, err := fileHash.Write(fileData); err != nil {
		log.Fatalf("Unable to write data to hash: %v\n", err)
	}

	chkSumBytes, err := fmt.Fprintf(chkSumFileHandle, "%x %s", fileHash.Sum(nil), driveFileId)
	if err != nil {
		log.Fatalf("Unable to write data to checksum file: %v\n", err)
	}
	fmt.Printf("Writed %v bytes to checksum file\n", chkSumBytes)

}

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync [file|dir]",
	Short: "Sync a file or a directory",
	Long: `Sync/backup a file or a directory:
"dsync sync [file|dir]"
If a directory is specified it will be synced recurrently.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		//Time benchmarking
		startTime := time.Now()
		defer func() {
			fmt.Printf("Time enlapsed: %v\n", time.Since(startTime))
		}()
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

		fileToSync, err := filepath.Abs(args[0])
		if err != nil {
			log.Fatalf("Unable to get file or directory %q: %v", args[0], err)
		}
		fileStats, err := os.Lstat(fileToSync)
		if err != nil {
			log.Fatalf("Unable to get file or dir %q stats: %v", args[0], err)
		}

		switch {
		case fileStats.Mode().IsDir():
			SyncDir(fileToSync, nil, srv)

		case fileStats.Mode().IsRegular():
			SyncFile(fileToSync, nil, srv)
		}

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
}
