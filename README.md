# pastego

Scrape pastebin with API using GO

## Usage

Search keywords are case sensitive

`pastego -s "password,keygen,PASSWORD"`


```
usage: pastego [<flags>]

Flags:
      --help              Show context-sensitive help (also try --help-long and --help-man).
  -s, --search="pass"     Strings to search, i.e: "password,ssh"
  -o, --output="results"  Folder to save the bins
  -i, --insensitive       Search for case-insensitive strings
```

## Requirements

`go get -u "github.com/PuerkitoBio/goquery"`

`go get -u "gopkg.in/alecthomas/kingpin.v2"`

## Disclaimer

You need a PRO account to use this: pastebin will **block/blacklist** your IP.

[pastebin PRO](https://pastebin.com/pro)

##### Or....

- increase the time between each request
- create a script to restart your router when pastebin warns you
