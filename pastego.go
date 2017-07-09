package main

import (
	// import standard libraries
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/edoz90/pastego/gui"
	"github.com/edoz90/pastego/pegmatch"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

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

func pasteSearcher(link *pasteJSON) {
	doc, err := goquery.NewDocument(link.ScrapeURL)
	if err != nil {
		log.Fatalln(err)
	}
	doc.Find("body").Each(func(index int, item *goquery.Selection) {
		if res, match := contains(item.Text(), strings.Split(*searchFor, ",")); res {
			if saveToFile(link, item.Text(), match) {
				s := fmt.Sprintf("%s - %s", match, link.FullURL)
				gui.PrintTo("log", s)
				gui.ListDir()
			}
		}
	})
}

func getBins(bins int) []pasteJSON {
	url := "https://pastebin.com/api_scraping.php?limit=" + fmt.Sprint(bins)
	slowDown := "Please slow down"
	trans := &http.Transport{DisableKeepAlives: false}
	client := &http.Client{Timeout: 60 * time.Second, Transport: trans}
	var out []pasteJSON

	r, err := client.Get(url)
	if err != nil {
		log.Fatalln(err)
	}
	if r != nil {
		defer r.Body.Close()
	}
	// read []byte{}
	b, _ := ioutil.ReadAll(r.Body)

	// Due to some presence of unicode chars convert raw JSON to string than parse it
	// GO strings works with utf-8
	if err = json.NewDecoder(strings.NewReader(string(b))).Decode(&out); err != nil {
		if strings.Contains(string(b), slowDown) || string(b) == "" {
			gui.PrintTo("log", "Slow down!\n")
		} else {
			s := fmt.Sprintf("%s\n", string(b))
			gui.PrintTo("log", s)
			log.Fatalln(err)
		}
	}
	return out
}

func saveToFile(link *pasteJSON, text string, match string) bool {
	// ./outputDir
	outputDir, _ := filepath.Abs(filepath.Clean(*outputTo))
	if err := os.MkdirAll(outputDir, os.FileMode(0775)); err != nil {
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
	gui.PrintTo("log", "Done!\n")

	// Run every 'interval' seconds
	for range time.NewTicker(time.Duration(interval) * time.Second).C {
		gui.PrintTo("log", "Restarting...")
		parseBins()
		gui.PrintTo("log", "Done!\n")
	}
}

func main() {
	kingpin.Parse()
	// Without a PRO account try to increase the first args and decrease the second
	go run(150, 250)
	gui.SetGui(*outputTo)
}
