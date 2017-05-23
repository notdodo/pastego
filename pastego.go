package main

import (
	// import standard libraries
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	// import third party libraries
	"github.com/PuerkitoBio/goquery"
	"gopkg.in/alecthomas/kingpin.v2"
)

var paste string = "https://pastebin.com"
var links []string

// Command line args
var (
	searchFor  = kingpin.Flag("search", "Strings to search, i.e: \"password,ssh\"").Short('s').Default("pass").String()
	outputTo   = kingpin.Flag("output", "Folder to save the bins").Short('o').Default("results").String()
	caseInsens = kingpin.Flag("insensitive", "Search for case-insensitive strings").Default("false").Short('i').Bool()
)

type pasteJSON struct {
	ScrapeURL string `json:"scrape_url"`
	FullURL   string `json:"full_url"`
	Date      string `json:"date"`
	Key       string `json:"key"`
	Size      string `json:"size"`
	Expire    string `json:"expire"`
	Title     string `json:"title"`
	Syntax    string `json:"syntax"`
	User      string `json:"user"`
}

func contains(link string, cleaners []string) (bool, string) {
	for _, clr := range cleaners {
		if *caseInsens {
			if strings.EqualFold(link, clr) {
				return true, clr
			}
		} else {
			if strings.Contains(link, clr) {
				return true, clr
			}

		}
	}
	return false, ""
}

func pasteSearcher(link *pasteJSON) {
	doc, err := goquery.NewDocument(link.ScrapeURL)
	if err != nil {
		fmt.Printf("searcher\n\n")
		log.Fatal(err)
	}
	doc.Find("body").Each(func(index int, item *goquery.Selection) {
		if res, match := contains(item.Text(), strings.Split(*searchFor, ",")); res {
			fmt.Printf("%s - %s\n", match, link.FullURL)
			saveToFile(link, item.Text(), match)
		}
	})
}

func getBins(bins int) []pasteJSON {
	var url string = "https://pastebin.com/api_scraping.php?limit=" + fmt.Sprint(bins)
	var slowDown string = "Please slow down"
	var trans = &http.Transport{DisableKeepAlives: false}
	var client = &http.Client{Timeout: 5 * time.Second, Transport: trans}
	var out []pasteJSON

	r, err := client.Get(url)
	if err != nil {
		fmt.Printf("get\n\n")
		log.Fatal(err)
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
			fmt.Printf("Slow down!\n\n")
			time.Sleep(10 * time.Second)
		} else {
			fmt.Printf("jsondecoder\n\n")
			log.Fatal(err)
		}
	}
	return out
}

func saveToFile(link *pasteJSON, text string, match string) {
	os.Mkdir(*outputTo, os.FileMode(0775))
	var title string = fmt.Sprintf("%s/%s - ", *outputTo, match)
	if link.Title == "" {
		title += link.Key
	} else {
		title += link.Title
	}
	if _, err := os.Stat(title); os.IsNotExist(err) {
		if err := ioutil.WriteFile(title, []byte(text), 0644); err != nil {
			fmt.Printf("writefile!\n\n")
			log.Fatal(err)
		}
	}
}

func run(interval int, bins int) {
	for _, v := range getBins(bins) {
		pasteSearcher(&v)
	}
	for range time.NewTicker(time.Duration(interval) * time.Second).C {
		fmt.Printf("Restarting...\n")
		for _, v := range getBins(bins) {
			pasteSearcher(&v)
		}
		fmt.Printf("Done!\n\n")
	}
}

func main() {
	kingpin.Parse()
	run(120, 250)
}
