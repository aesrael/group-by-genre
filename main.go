package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/dhowden/tag"
)

const (
	Genres     = "Genres"
	NoGenre    = "No Genre"
	Duplicates = "Duplicate"
	MusicPath  = "Music"
)

func main() {
	printNoticeAndUsageInstruction()
	// Get the user's home directory
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err, usr)
	}

	defaultMusicLibraryPath := filepath.Join(usr.HomeDir, MusicPath)

	reader := bufio.NewReader(os.Stdin)

	// Prompt for music library path
	fmt.Printf("\n\033[31mEnter the music library path (leave empty for default %s): \033[0m\n", defaultMusicLibraryPath)
	musicLibraryPath, _ := reader.ReadString('\n')
	musicLibraryPath = strings.TrimSpace(musicLibraryPath)
	if musicLibraryPath == "" {
		musicLibraryPath = defaultMusicLibraryPath
	}

	genresFolderPath := musicLibraryPath + "/" + Genres
	noGenreFolderPath := filepath.Join(genresFolderPath, NoGenre)

	err = filepath.WalkDir(musicLibraryPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() && d.Name() == Genres {
			return filepath.SkipDir // Skip the "genres" directory
		}

		extension := strings.ToLower(filepath.Ext(path))
		if !d.IsDir() && !strings.HasPrefix(d.Name(), ".") && (extension == ".mp3" || extension == ".flac") {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			metadata, err := tag.ReadFrom(file)
			if err != nil {
				return err
			}

			genre := capitalize(metadata.Genre())
			var destinationFolder string

			if genre != "" {
				destinationFolder = filepath.Join(genresFolderPath, genre)
			} else {
				destinationFolder = noGenreFolderPath
			}
			sourceFolder := filepath.Dir(path)

			// Create the destination folder if it doesn't exist
			if _, err := os.Stat(destinationFolder); os.IsNotExist(err) {
				os.MkdirAll(destinationFolder, os.ModePerm)
			}

			destinationPath := filepath.Join(destinationFolder, d.Name())

			// Check if the song already exists in the destination folder
			_, err = fs.Stat(os.DirFS(destinationFolder), d.Name())
			switch {
			case err == nil:
				// Song already exists, move it to the "duplicate" folder
				duplicateFolder := filepath.Join(genresFolderPath, Duplicates)
				duplicatePath := filepath.Join(duplicateFolder, d.Name())

				// Create the "duplicate" folder if it doesn't exist
				if _, err := os.Stat(duplicateFolder); os.IsNotExist(err) {
					os.MkdirAll(duplicateFolder, os.ModePerm)
				}

				// Move the song to the "duplicate" folder
				err = os.Rename(path, duplicatePath)
				if err != nil {
					return err
				}

				fmt.Printf("Song %s already exists in the destination folder. Moved to %s\n", d.Name(), duplicateFolder)
			case errors.Is(err, fs.ErrNotExist):
				// Song doesn't exist, move it to the destination folder
				err = os.Rename(path, destinationPath)
				if err != nil {
					return err
				}

				fmt.Printf("Moved %s from %s to %s\n", sourceFolder, d.Name(), destinationFolder)
			default:
				return err
			}
		}

		return nil
	})

	if err != nil {
		log.Fatal("Error:", err)
	}
}

func capitalize(s string) string {
	return strings.Title(s)
}

func printNoticeAndUsageInstruction() {
	// add app name here such as fmt.Print("Music Genre Organizer\n") but wuth some color
	fmt.Printf("\n\033[31m 🎶Music Genre Organizer:\033[0m\n\n")
	fmt.Printf("\033[1;35mNotice and Usage Instructions:\033[0m\n")
	fmt.Printf("This program organizes a music library by grouping songs into genre-based folders.\n")
	fmt.Printf("currently only supports MP3 & flac files\n")
	fmt.Printf("Songs with no genre information are moved to a separate 'No Genre' folder.\n")
	fmt.Printf("Duplicate songs found in the destination folders are moved to a 'Duplicate' folder.\n")
	fmt.Printf("\n")
	fmt.Printf("\033[1;35mUsage Instructions:\033[0m\n")
	fmt.Printf("Please follow the steps below to use the program:\n")
	fmt.Printf("1. Provide the path to the music library (defaults to $Home/Music) directory when prompted.\n")
	fmt.Printf("3. The program will scan the music library, group songs by genre, and move them to respective genre folders.\n")
	fmt.Printf("4. Songs without a genre will be moved to the 'No Genre' folder.\n")
	fmt.Printf("5. If duplicate songs are found, they will be moved to the 'Duplicate' folder.\n")
	fmt.Print("*************************************************************************************************\n\n")
}
