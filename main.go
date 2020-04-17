// ccpc crypto coin price checker

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"time"

	"../ccpc/cgapi"
	"github.com/gookit/color"
)

const (
	userAgent string = "ccpc"
)

// target is used to set the currency for comparison.
type target struct {
	id string
}

// Listing defines included elements in a possible listing.
type listing struct {
	symbol             bool
	name               bool
	blockTimeInMinutes bool
	lastUpdated        bool
	volume             bool
	target             string
	color              bool
}

// DefaultListing defines the standard set of included elements in a listing.
func defaultListing() listing {
	self := listing{}
	self.symbol = true
	self.name = true
	self.target = "USD"
	self.lastUpdated = true
	self.color = true
	return self
}

// MaxListing defines the maximal set of included elements in a listing.
func maxListing() listing {
	self := listing{}
	self.symbol = true
	self.name = true
	self.target = "USD"
	self.lastUpdated = true
	self.color = true
	self.blockTimeInMinutes = true
	self.volume = true
	return self
}

// Entry point handles Args and flags
func main() {
	// CLI flag handling
	allPtr := flag.Bool("all", false, "Yields default listings for all supported coins.")
	flag.Parse()

	// -all
	if *allPtr {
		var keys = make([]string, len(cgapi.CGCoinURLs))
		i := 0
		for k := range cgapi.CGCoinURLs {
			keys[i] = k
			i++
		}
		sort.Strings(keys)
		for key := 0; key < len(keys); key++ {
			res, _ := httpRequest(cgapi.CGCoinURLs[keys[key]], userAgent)
			var coin cgapi.CGCoinSingleton
			json.Unmarshal(res, &coin)
			displayCoinListing(coin, defaultListing(), true)
		}
	}

	// res, _ := httpRequest(cgapi.CGCoinURLs["btc"], userAgent)
	// var coin cgapi.CGCoinSingleton
	// json.Unmarshal(res, &coin)
	// var testlisting listing
	// testlisting.color = true
	// testlisting.name = true
	// testlisting.target = "JPY"
	// testlisting.volume = true
	// testlisting.lastUpdated = true
	// testlisting.blockTimeInMinutes = true
	// displayCoinListing(coin, testlisting, false)
}

// Display a coin ticker.
func displayCoinListing(coin cgapi.CGCoinSingleton, list listing, leftAlign bool) {
	if len(coin.Symbol) > 1 {
		buf := new(bytes.Buffer) // buffer is only used when not in color mode
		if list.color {
			color.New(color.FgBlack, color.BgBlue).Print(ctir(coin.Symbol, 5, leftAlign))
		} else {
			buf.WriteString(ctir(coin.Symbol, 5, leftAlign))
		}
		if list.name {
			if list.color {
				color.FgBlue.Print(ctir(coin.Name, 15, leftAlign))
			} else {
				buf.WriteString(fmt.Sprint(ctir(coin.Name, 15, leftAlign)))
			}
		}
		for i, t := range coin.Tickers {
			if t.Target == list.target {
				if list.color {
					if coin.MarketData.PriceChange24h >= 0 {
						color.BgGreen.Print(ctir(cgapi.MonetarySymbols[list.target]+
							fmt.Sprintf("%.2f", coin.Tickers[i].Last), 14, leftAlign))
					} else {
						color.BgRed.Print(ctir(cgapi.MonetarySymbols[list.target]+
							fmt.Sprintf("%.2f", coin.Tickers[i].Last), 14, leftAlign))
					}
				} else {
					buf.WriteString(ctir(cgapi.MonetarySymbols[list.target]+
						fmt.Sprintf("%.2f", coin.Tickers[i].Last), 14, leftAlign))
				}
				if list.volume {
					if list.color {
						color.BgDarkGray.Print(ctir("VOL:"+fmt.Sprintf("%.8f", coin.Tickers[i].Volume), 25, leftAlign))
					} else {
						buf.WriteString(ctir("VOL:"+fmt.Sprintf("%.8f", coin.Tickers[i].Volume), 25, leftAlign))
					}
				}
				break
			}
			if i == len(coin.Tickers)-1 {
				buf.WriteString("N/A ")
			}
		}
		if list.lastUpdated || list.blockTimeInMinutes {
			if list.lastUpdated {
				tm, _ := time.Parse(time.RFC3339Nano, coin.LastUpdated)
				if list.color {
					color.BgDarkGray.Print(ctir("UPD:"+tm.Format(time.RFC822), 25))
				} else {
					buf.WriteString(ctir("UPD:"+tm.Format(time.RFC822), 25))
				}
			}
			if list.blockTimeInMinutes {
				if list.color {
					color.BgDarkGray.Print(ctir("BT:"+fmt.Sprintf("%.2f", coin.BlockTimeInMinutes)+" MINS", 16, leftAlign))
				} else {
					buf.WriteString(ctir("BT:"+fmt.Sprintf("%.2f", coin.BlockTimeInMinutes)+" MINS", 16, leftAlign))
				}
			}
		}
		if !list.color {
			buf.WriteString(" ")
			fmt.Println(buf.String())
		} else {
			color.BgDarkGray.Println(" ")
		}
	} else {
		if list.color {
			color.BgYellow.Println(" Symbol not found or network error. ")
		} else {
			fmt.Println(" Symbol not found or network error. ")
		}
	}
}

// Performs an HTTP request.
func httpRequest(url, userAgent string) (contents []byte, err error) {
	cli := http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	res, err := cli.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	contents, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return
}

// Returns a string which is centered in the middle of the range.
func ctir(str string, rng int, left ...bool) string {
	if len(str) > rng {
		log.Fatal("String too long for range.")
	}
	diff := rng - len(str)
	buf := new(bytes.Buffer)
	if len(left) > 0 && left[0] && diff > 2 {
		buf.WriteString("  " + str)
		for i := 1 + len(str); i < rng; i++ {
			buf.WriteString(" ")
		}
	} else {
		if diff%2 == 0 {
			for i := 0; i < diff/2; i++ {
				buf.WriteString(" ")
			}
			buf.WriteString(str)
			for i := 0; i < diff/2; i++ {
				buf.WriteString(" ")
			}
		} else {
			for i := 0; i < (diff-1)/2; i++ {
				buf.WriteString(" ")
			}
			buf.WriteString(" " + str)
			for i := 0; i < (diff-1)/2; i++ {
				buf.WriteString(" ")
			}
		}
	}
	return buf.String()
}
