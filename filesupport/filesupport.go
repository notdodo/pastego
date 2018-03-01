package filesupport

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
)

type PasteJSON struct {
	ScrapeURL string `json:"scrape_url,-"`
	FullURL   string `json:"full_url,-"`
	Date      string `json:"date,-"`
	Key       string `json:"key,-"`
	Size      string `json:"size,-"`
	Expire    string `json:"expire,-"`
	Title     string `json:"title,-"`
	Syntax    string `json:"syntax,-"`
	User      string `json:"user,-"`
}

var logFile string

// Log a string to a temp file
func LogToFile(s string) {
	var err error
	var tmpfile *os.File
	var t = time.Now().Format(time.RFC3339)
	if logFile != "" {
		tmpfile, err = os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	} else {
		tmpfile, err = ioutil.TempFile("", "pastego")
		logFile = tmpfile.Name()
	}
	if err != nil {
		log.Fatalln(err)
	}

	if _, err := tmpfile.Write([]byte(t + " - " + s + "\r\n")); err != nil {
		log.Fatalln(err)
	}

	defer tmpfile.Close()
}

// Save the bin to the output directory: default is '$(pwd)/results'
func SaveToFile(link *PasteJSON, text string, match string, outputTo string) bool {
	// ./outputDir
	outputDir, _ := filepath.Abs(filepath.Clean(outputTo))
	if err := os.MkdirAll(outputDir, os.FileMode(0775)); err != nil {
		// Error on creating/reading the output folder
		LogToFile(err.Error())
		log.Fatalln(err)
	}
	// match - pasteTitle
	var title string
	if link.Title == "" {
		title += link.Key
	} else {
		title += link.Title
	}
	title = fmt.Sprintf("%s__", match) + govalidator.SafeFileName(strings.Replace(title, "/", "_", -1))
	// ./outputDir/match - pasteTitle
	filePath := outputDir + string(filepath.Separator) + title
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if err := ioutil.WriteFile(filePath, []byte(text), 0644); err != nil {
			// Error on writing file, something went wrong
			LogToFile(err.Error())
			log.Fatalln(err)
		}
		return true
	}
	return false
}

// Delete a file when is not interesting
func DeleteFile(l string, baseDir string) error {
	f, _ := filepath.Abs(baseDir + string(filepath.Separator) + l)
	if _, err := os.Stat(f); !os.IsNotExist(err) {
		if err := os.Remove(f); err != nil {
			return err
		}
	}
	return nil
}
