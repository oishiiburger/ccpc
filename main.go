// ccpc crypto coin price checker

package main

import (
	"bytes"
	"encoding/json"

	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"ccpc/cgapi"

	"github.com/gookit/color"
	flag "github.com/ogier/pflag"
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
	blockTimeInMinutes bool
	color              bool
	lastUpdated        bool
	name               bool
	symbol             bool
	target             string
	volume             bool
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
	var listingProps listing = defaultListing()

	// Set usage message
	flag.Usage = func() {
		fmt.Println("Usage: ccpc symbol(s) [options]")
		fmt.Println("Options:")
		flag.PrintDefaults()
	}

	// CLI flag handling
	allPtr := flag.BoolP("all", "a", false, "Yields listings for all supported coins.")
	blkPtr := flag.BoolP("block-time", "b", false, "Includes block time in the listing, if available.")
	bwtPtr := flag.BoolP("no-color", "c", false, "Disables output colors.")
	maxPtr := flag.BoolP("maximum", "m", false, "Yields maximum detail listings for the selected coins.")
	namPtr := flag.BoolP("no-name", "n", false, "Omits coin name in the listing.")
	tgtPtr := flag.StringP("target", "t", "usd", "Determines the target currency for comparison (e.g. usd, jpy).")
	timPtr := flag.BoolP("no-time", "z", false, "Omits last update time in the listing.")
	volPtr := flag.BoolP("volume", "v", false, "Includes coin volume in the listing, if available.")
	flag.Parse()
	// maxListing is copied over listingProperties, so it must be first
	if *maxPtr {
		listingProps = maxListing()
	}
	if *blkPtr {
		listingProps.blockTimeInMinutes = true
	}
	if *bwtPtr {
		listingProps.color = false
	}
	if *namPtr {
		listingProps.name = false
	}
	if len(*tgtPtr) > 0 {
		listingProps.target = strings.ToUpper(*tgtPtr)
	}
	if *timPtr {
		listingProps.lastUpdated = false
	}
	if *volPtr {
		listingProps.volume = true
	}
	// --all needs other listingProperties ready
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
			displayCoinListing(coin, listingProps, true)
		}
	}

	// CLI Args handling
	if len(os.Args) == 1 {
		flag.Usage()
	} else if !*allPtr {
		var args []string
		for _, arg := range os.Args[1:] {
			if !strings.Contains(arg, "-") && len(cgapi.CGCoinURLs[arg]) > 0 {
				args = append(args, arg)
			}
		}
		for _, arg := range args {
			var leftAlign bool = false
			if len(args) > 1 {
				leftAlign = true // align if there are multiple symbols to check
			}
			res, _ := httpRequest(cgapi.CGCoinURLs[arg], userAgent)
			var coin cgapi.CGCoinSingleton
			json.Unmarshal(res, &coin)
			displayCoinListing(coin, listingProps, leftAlign)
		}
	}
}

// Display a coin ticker.
func displayCoinListing(coin cgapi.CGCoinSingleton, list listing, leftAlign bool) {
	if len(coin.Symbol) > 1 {
		buf := new(bytes.Buffer) // buffer is only used when not in color mode
		if list.color {
			color.New(color.FgBlack, color.BgBlue).Print(ctir(coin.Symbol, 6, leftAlign))
		} else {
			buf.WriteString(ctir(coin.Symbol, 6, leftAlign))
		}
		if list.name {
			if list.color {
				color.FgBlue.Print(ctir(coin.Name, 16, leftAlign))
			} else {
				buf.WriteString(fmt.Sprint(ctir(coin.Name, 16, leftAlign)))
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
				if list.color {
					color.BgDarkGray.Print(ctir(cgapi.MonetarySymbols[list.target]+"------", 14, leftAlign))
				} else {
					buf.WriteString(ctir(cgapi.MonetarySymbols[list.target]+"------", 14, leftAlign))
				}
				if list.volume {
					if list.color {
						color.BgDarkGray.Print(ctir("", 25, leftAlign))
					} else {
						buf.WriteString(ctir("", 25, leftAlign))
					}
				}
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
					color.BgDarkGray.Print(ctir("BT:"+fmt.Sprintf("%.1f", coin.BlockTimeInMinutes)+"m", 12, leftAlign))
				} else {
					buf.WriteString(ctir("BT:"+fmt.Sprintf("%.1f", coin.BlockTimeInMinutes)+"m", 12, leftAlign))
				}
			}
		}
		if !list.color {
			buf.WriteString(" ")
			fmt.Print(buf.String())
		} else {
			color.BgDarkGray.Print(" ")
		}
	} else {
		if list.color {
			color.BgYellow.Print(ctir("Symbol not found or network error.", 38, leftAlign))
		} else {
			fmt.Print(ctir("Symbol not found or network error.", 38, leftAlign))
		}
	}
	fmt.Println()
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
		log.Fatal("String too long for range when centering in column.")
	}
	if str == "" {
		str = "(unknown)"
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
			buf.WriteString(str + " ")
			for i := 0; i < (diff-1)/2; i++ {
				buf.WriteString(" ")
			}
		}
	}
	return buf.String()
}
