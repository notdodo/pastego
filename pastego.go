package main

import (
	// import standard libraries
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/edoz90/pastego/filesupport"
	"github.com/edoz90/pastego/gui"
	"github.com/edoz90/pastego/pegmatch"

	// import third party libraries
	"github.com/PuerkitoBio/goquery"
	"gopkg.in/alecthomas/kingpin.v2"
)

// Command line args
var (
	searchFor  = kingpin.Flag("search", "Strings to search with optional bool operator(&&, ||, ~), i.e: \"password,some || (thing && ~maybenot), \"").Short('s').Default("pass").String()
	outputTo   = kingpin.Flag("output", "Folder to save the bins. Default : './results'").Short('o').Default("results").String()
	caseInsens = kingpin.Flag("insensitive", "Search for case-insensitive strings").Default("false").Short('i').Bool()
)

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

// Parse the page and read the content of the bin
func pasteSearcher(link *filesupport.PasteJSON) {
	client := &http.Client{Timeout: 10 * time.Second}
	response, err := client.Get(link.ScrapeURL)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		log.Fatalln(err)
	}
	doc.Find("body").Each(func(index int, item *goquery.Selection) {
		if res, match := contains(item.Text(), strings.Split(*searchFor, ",")); res {
			if filesupport.SaveToFile(link, item.Text(), match, *outputTo) {
				var s string
				if link.Title != "" {
					s = fmt.Sprintf("%s - %s - %s", match, link.FullURL, link.Title)
				} else {
					s = fmt.Sprintf("%s - %s", match, link.FullURL)
				}
				// Show recent pastes
				gui.PrintTo("log", s)
				logToFile(s)
				// Triggers a reload
				gui.ListDir()
			}
		}
	})
}

// Fetch the bins
func getBins(bins int) []filesupport.PasteJSON {
	url := "https://scrape.pastebin.com/api_scraping.php?limit=" + fmt.Sprint(bins)
	slowDown := "Please slow down"
	client := &http.Client{Timeout: 10 * time.Second}
	var out []filesupport.PasteJSON

	r, err := client.Get(url)
	if err != nil {
		logToFile(err.Error())
		return out
	}
	defer r.Body.Close()
	if r != nil {
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

// Set the program to fetch `bins` bins every `interval` seconds
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

// Wrapper to avoid writing log function calls :)
func logToFile(s string) {
	filesupport.LogToFile(s)
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

	// Without a PRO account try to increase the first args and decrease the second.
	go run(150, 250)
	gui.SetGui(*outputTo)
}
