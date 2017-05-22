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
var cleaners = []string{"/pro", "/scraping", "/archive", "/deals/", "/trends", "/api", "/tools", "/faq", "/login", "/messages", "/alerts", "/settings"}
var links []string

var (
	searchFor = kingpin.Flag("search", "Strings to search").Short('s').Default("pass").String()
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

func contains(link string, cleaners []string) bool {
	for _, clr := range cleaners {
		if strings.Contains(link, clr) {
			return true
		}
	}
	return false
}

func pasteCollector() {
	doc, err := goquery.NewDocument(paste + "/archive")
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("a").Each(func(index int, item *goquery.Selection) {
		linkTag := item
		link, _ := linkTag.Attr("href")
		if !contains(link, cleaners) {
			links = append(links, link)
		}
	})
}

func pasteSearcher(link *pasteJSON) {
	doc, err := goquery.NewDocument(link.ScrapeURL)
	if err != nil {
		fmt.Printf("searcher\n\n")
		log.Fatal(err)
	}
	doc.Find("body").Each(func(index int, item *goquery.Selection) {
		if contains(item.Text(), strings.Split(*searchFor, ",")) {
			fmt.Printf("%s\n", link.FullURL)
			saveToFile(link, item.Text())
		}
	})
}

func getBins() []pasteJSON {
	var url string = "https://pastebin.com/api_scraping.php?limit=250"
	var slowDown string = "Please slow down"
	var client = &http.Client{Timeout: 5 * time.Second}
	var out []pasteJSON

	r, err := client.Get(url)
	if err != nil {
		fmt.Printf("get\n\n")
		log.Fatal(err)
	}
	defer r.Body.Close()
	// read []byte{}
	b, _ := ioutil.ReadAll(r.Body)

	// Due to some presence of unicode char converto JSON to string than parse it
	// Go strings works with utf-8
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

func saveToFile(link *pasteJSON, text string) {
	_ = os.Mkdir("results", os.FileMode(0775))
	var title string = "results/"
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

func run(interval int) {
	for _, v := range getBins() {
		pasteSearcher(&v)
	}
	for range time.NewTicker(time.Duration(interval) * time.Second).C {
		fmt.Printf("Restarting...\n")
		for _, v := range getBins() {
			pasteSearcher(&v)
		}
		fmt.Printf("Done!\n\n")
	}
}

func main() {
	kingpin.Parse()
	run(120)
}
