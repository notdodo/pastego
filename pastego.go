package main

import (
	// import standard libraries
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/edoz90/pastego/gui"
	"github.com/edoz90/pastego/pegmatch"

	// import third party libraries
	"github.com/PuerkitoBio/goquery"
	"github.com/asaskevich/govalidator"
	"gopkg.in/alecthomas/kingpin.v2"
)

// Command line args
var (
	searchFor  = kingpin.Flag("search", "Strings to search with optional bool operator(&&, ||, ~), i.e: \"password,some || (thing && ~maybenot), \"").Short('s').Default("pass").String()
	outputTo   = kingpin.Flag("output", "Folder to save the bins").Short('o').Default("results").String()
	caseInsens = kingpin.Flag("insensitive", "Search for case-insensitive strings").Default("false").Short('i').Bool()
)

type pasteJSON struct {
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

var logFile string = ""

// Using PEG check if the bin contains the searched word/s
func contains(link string, matches []string) (bool, string) {
	var origMtch = make([]string, len(matches))
	copy(origMtch, matches)
	if *caseInsens {
		link = strings.ToUpper(link)
		for i, v := range matches {
			matches[i] = strings.ToUpper(v)
		}
	}
	pegmatch.PasteContentString = link
	for i, mtch := range matches {
		mtch = strings.TrimSpace(mtch)
		got, err := pegmatch.ParseReader("", bytes.NewBufferString(mtch))
		if err == nil && got.(bool) {
			return true, strings.Split(origMtch[i], " ")[0]
		}
	}
	return false, ""
}

// Read the bin
func pasteSearcher(link *pasteJSON) {
	doc, err := goquery.NewDocument(link.ScrapeURL)
	if err != nil {
		log.Fatalln(err)
	}
	doc.Find("body").Each(func(index int, item *goquery.Selection) {
		if res, match := contains(item.Text(), strings.Split(*searchFor, ",")); res {
			if saveToFile(link, item.Text(), match) {
				var s string
				if link.Title != "" {
					s = fmt.Sprintf("%s - %s - %s", match, link.FullURL, link.Title)
				} else {
					s = fmt.Sprintf("%s - %s", match, link.FullURL)
				}
				// Show recent pastes
				gui.PrintTo("log", s)
				logToFile(s)
				gui.ListDir()
			}
		}
	})
}

// Get a list of 'bins' bin
func getBins(bins int) []pasteJSON {
	url := "https://pastebin.com/api_scraping.php?limit=" + fmt.Sprint(bins)
	slowDown := "Please slow down"
	trans := &http.Transport{DisableKeepAlives: false}
	client := &http.Client{Timeout: 60 * time.Second, Transport: trans}
	var out []pasteJSON

	r, err := client.Get(url)
	if err != nil {
		logToFile(err.Error())
	}
	if r != nil {
		defer r.Body.Close()
		// read []byte{}
		b, _ := ioutil.ReadAll(r.Body)

		// Due to some presence of unicode chars convert raw JSON to string than parse it
		// GO strings works with utf-8
		if err = json.NewDecoder(strings.NewReader(string(b))).Decode(&out); err != nil {
			if strings.Contains(string(b), slowDown) || string(b) == "" {
				logToFile("Slow down!\n")
			} else {
				// Error on marshalling JSON
				s := fmt.Sprintf("\n%s\n", string(b))
				logToFile(s)
			}
		}
	}
	return out
}

func logToFile(s string) {
	var err error
	var tmpfile *os.File
	if logFile != "" {
		tmpfile, err = os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	} else {
		tmpfile, err = ioutil.TempFile("", "pastego")
		logFile = tmpfile.Name()
	}
	if err != nil {
		log.Fatalln(err)
	}

	if _, err := tmpfile.Write([]byte(s + "\r\n")); err != nil {
		log.Fatalln(err)
	}

	defer tmpfile.Close()
}

func saveToFile(link *pasteJSON, text string, match string) bool {
	// ./outputDir
	outputDir, _ := filepath.Abs(filepath.Clean(*outputTo))
	if err := os.MkdirAll(outputDir, os.FileMode(0775)); err != nil {
		// Error on creating/reading the output folder
		logToFile(err.Error())
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
			logToFile(err.Error())
			log.Fatalln(err)
		}
		return true
	}
	return false
}

func run(interval int, bins int) {
	parseBins := func() {
		for _, v := range getBins(bins) {
			pasteSearcher(&v)
		}
	}

	// First run
	parseBins()
	logToFile("Done!\n")

	// Run every 'interval' seconds
	for range time.NewTicker(time.Duration(interval) * time.Second).C {
		logToFile("Restarting...")
		parseBins()
		logToFile("Done!\n")
	}
}

func main() {
	kingpin.Parse()
	logToFile(`

		██████╗  █████╗ ███████╗████████╗███████╗ ██████╗  ██████╗ 
		██╔══██╗██╔══██╗██╔════╝╚══██╔══╝██╔════╝██╔════╝ ██╔═══██╗
		██████╔╝███████║███████╗   ██║   █████╗  ██║  ███╗██║   ██║
		██╔═══╝ ██╔══██║╚════██║   ██║   ██╔══╝  ██║   ██║██║   ██║
		██║     ██║  ██║███████║   ██║   ███████╗╚██████╔╝╚██████╔╝
		╚═╝     ╚═╝  ╚═╝╚══════╝   ╚═╝   ╚══════╝ ╚═════╝  ╚═════╝ 	
	`)

	// Without a PRO account try to increase the first args and decrease the second
	go run(150, 250)
	gui.SetGui(*outputTo)
}
