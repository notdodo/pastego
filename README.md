# pastego [![Build Status](https://travis-ci.org/edoz90/pastego.svg?branch=master)](https://travis-ci.org/edoz90/pastego) <a href="https://www.buymeacoffee.com/d0d0" target="_blank"><img src="https://www.buymeacoffee.com/assets/img/custom_images/yellow_img.png" alt="Buy Me A Coffee" style="height: 41px !important;width: 174px !important;box-shadow: 0px 3px 2px 0px rgba(190, 190, 190, 0.5) !important;-webkit-box-shadow: 0px 3px 2px 0px rgba(190, 190, 190, 0.5) !important;" ></a>

Scrape/Parse Pastebin using GO and grammar expression (PEG).
                                                         
![pastego.png](https://raw.githubusercontent.com/edoz90/pastego/support/pastego.png)


## Installation

`$ go get -u github.com/edoz90/pastego`

## Usage

Search keywords are case sensitive

`pastego -s "password,keygen,PASSWORD"`

You can use boolean operators to reduce false positive

`pastego -s "quake && ~earthquake, password && ~(php || sudo || Linux || '<body>')"`

This command will search for bins with `quake` but not `earthquake` words and for bins with `password` but not `php`, `sudo`, `Linux`, `<body>` words.

```
usage: pastego [<flags>]

Flags:
      --help              Show context-sensitive help (also try --help-long and --help-man).
  -s, --search="pass"     Strings to search, i.e: "password,ssh"
  -o, --output="results"  Folder to save the bins
  -i, --insensitive       Search for case-insensitive strings
```

Supported expression/operators:

    `&&` - and

    `||` - or

    `~` - not

    `'string with space'`

    `(myexpression && 'with operators')`

### Keybindings

`q`, `ctrl+c`: quit `pastego`

`k`, `↑`: show previous bin

`j`, `↓`: show next bin

`n`: jump forward by 15 bins

`p`: jump backward by 15 bins

`N`: move to the next block of findings (in alphabet order)

`P`: move to the previous block of findings (in alphabet order)

`d`: delete file from file system

`HOME`: go to top

## Requirements

#### [goquery](https://github.com/PuerkitoBio/goquery)

`go get -u "github.com/PuerkitoBio/goquery"`

#### [kingpin](https://github.com/alecthomas/kingpin)

`go get -u "gopkg.in/alecthomas/kingpin.v2"`

#### [gocui](https://github.com/jroimartin/gocui)

`go get -u "github.com/jroimartin/gocui"`

To create the code from PEG use [pigeon](https://github.com/mna/pigeon):

`go get -u github.com/mna/pigeon`

## Disclaimer

You need a PRO account to use this: pastebin will **block/blacklist** your IP.

[pastebin PRO](https://pastebin.com/pro)

#### Or....

- increase the time between each request
- create a script to restart your router when pastebin warns you

#### In progress...

Add flag to pass/read a list of proxies to avoid IP ban/throttle for free users
