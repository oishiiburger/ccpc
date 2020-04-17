// ccpc crypto coin price checker

package main

import (
	"bytes"
	"encoding/json"
	"log"
	"time"
	"unicode/utf8"

	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strings"

	"ccpc/cgapi"

	"github.com/gookit/color"
	flag "github.com/ogier/pflag"
)

const (
	userAgent string = "ccpc, https://github.com/oishiiburger/ccpc"
)

// target is used to set the currency for comparison.
type target struct {
	id string
}

// Listing defines included elements in a possible listing.
type listing struct {
	blockTIM         bool
	blockTIMWidth    int
	color            bool
	errWidth         int
	lastUpdated      bool
	lastUpdatedWidth int
	name             bool
	nameWidth        int
	priceWidth       int
	symbol           bool
	symbolWidth      int
	target           string
	volume           bool
	volumeWidth      int
}

// DefaultListingWidths defines the default field widths for a listing.
func defaultListingWidths() listing {
	self := listing{}
	self.blockTIMWidth = 11
	self.errWidth = 38
	self.lastUpdatedWidth = 27
	self.nameWidth = 25
	self.priceWidth = 28
	self.symbolWidth = 9
	self.volumeWidth = 18
	return self
}

// DefaultListing defines the standard set of included elements in a listing.
func defaultListing() listing {
	self := defaultListingWidths()
	self.name = true
	self.symbol = true
	self.target = "USD"
	self.lastUpdated = true
	self.color = true
	return self
}

// MaxListing defines the maximal set of included elements in a listing.
func maxListing() listing {
	self := defaultListing()
	self.symbol = true
	self.name = true
	self.target = "USD"
	self.lastUpdated = true
	self.color = true
	self.blockTIM = true
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
	allPtr := flag.BoolP("all", "a", false, "Yields listings for all known coins. (Generally not recommended)")
	blkPtr := flag.BoolP("block-time", "b", false, "Includes block time in the listing, if available.")
	bwtPtr := flag.BoolP("no-color", "c", false, "Disables output colors.")
	maxPtr := flag.BoolP("maximum", "m", false, "Yields maximum detail listings for the selected coins.")
	namPtr := flag.BoolP("no-name", "n", false, "Omits coin name in the listing.")
	tgtPtr := flag.StringP("target", "t", "usd", "Determines the target currency for comparison (e.g. usd, jpy).")
	timPtr := flag.BoolP("no-time", "z", false, "Omits last update time in the listing.")
	volPtr := flag.BoolP("volume", "v", false, "Includes coin volume in the listing, if available.")
	lcPtr := flag.Bool("list-coins", false, "Displays a listing of all known coins.")
	lmPtr := flag.Bool("list-currencies", false, "Displays a listing of all known currencies.")
	flag.Parse()
	// maxListing is copied over listingProperties, so it must be first
	if *maxPtr {
		listingProps = maxListing()
	}
	if *blkPtr {
		listingProps.blockTIM = true
	}
	if *bwtPtr {
		listingProps.color = false
	}
	if *lcPtr {
		listTableKeys(cgapi.CGCoinURLs, "coins")
	}
	if *lmPtr {
		listTableKeys(cgapi.MonetarySymbols, "currencies")
	}
	if *namPtr {
		listingProps.name = false
	}
	if len(*tgtPtr) > 0 {
		tgt := strings.ToUpper(*tgtPtr)
		if len(cgapi.MonetarySymbols[tgt]) > 0 {
			listingProps.target = tgt
		} else {
			errMessage("Unknown target currency: "+tgt+".", false, listingProps)
		}
	}
	if *timPtr {
		listingProps.lastUpdated = false
	}
	if *volPtr {
		listingProps.volume = true
	}
	// --all needs other listingProperties ready
	if *allPtr {
		keys := mapToSortedStrings(cgapi.CGCoinURLs)
		for key := 0; key < len(keys); key++ {
			res, _ := httpRequest(cgapi.CGCoinURLs[keys[key]], userAgent)
			var coin cgapi.CGCoinSingleton
			json.Unmarshal(res, &coin)
			generateCoinTicker(coin, listingProps)
		}
	}

	// CLI Args handling
	if len(os.Args) == 1 {
		flag.Usage()
	} else if !*allPtr {
		var args []string
		for _, arg := range os.Args[1:] {
			if !strings.Contains(arg, "-") &&
				(len(cgapi.CGCoinURLs[arg]) > 0 ||
					len(cgapi.CGCoinURLs[strings.ToUpper(arg)]) > 0) {
				args = append(args, arg)
			} else {
				errMessage("Coin symbol "+arg+" is not in the known set.", false, listingProps)
			}
		}
		for _, arg := range args {
			var symb = arg
			if cgapi.CGCoinURLs[symb] == "" {
				symb = strings.ToUpper(arg)
			}
			res, err := httpRequest(cgapi.CGCoinURLs[symb], userAgent)
			if err != nil {
				errMessage("HTTP request did not complete successfully.", true, listingProps)
			}
			var coin cgapi.CGCoinSingleton
			json.Unmarshal(res, &coin)
			generateCoinTicker(coin, listingProps)
		}
	}
}

// Generate a coin ticker.
func generateCoinTicker(coin cgapi.CGCoinSingleton, list listing) {
	var tickerIdx int
	if len(coin.Symbol) < 1 {
		// can this even happen?
		errMessage("Coin symbol was not successfully loaded.", true, list)
	} else {
		tPrint(coin.Symbol, list.symbol, list, color.BgBlue, list.symbolWidth)
		tPrint(coin.Name, list.name, list, color.FgBlue, list.nameWidth)
		if len(coin.Tickers) > 0 { // check to make sure this coin has a ticker
			for tickerIdx, t := range coin.Tickers {
				if t.Target == list.target {
					var per string
					if coin.MarketData.PriceChange24hPc != 0 {
						if coin.MarketData.PriceChange24hPc >= 0 {
							per = "+"
						}
						per = "(" + per + fmt.Sprintf("%3.2f", coin.MarketData.PriceChange24hPc) + "%/24h)"
					}
					if coin.MarketData.PriceChange24h >= 0 {
						tPrint(cgapi.MonetarySymbols[list.target]+fmt.Sprintf("%.2f", coin.Tickers[tickerIdx].Last)+
							" "+per, true, list, color.BgGreen, list.priceWidth)
					} else {
						tPrint(cgapi.MonetarySymbols[list.target]+fmt.Sprintf("%.2f", coin.Tickers[tickerIdx].Last)+
							" "+per, true, list, color.BgRed, list.priceWidth)
					}
					break
				}
				if tickerIdx == len(coin.Tickers)-1 {
					tPrint("no price", true, list, color.BgYellow, list.priceWidth)
				}
			}
		} else {
			tPrint("no price", true, list, color.BgYellow, list.priceWidth)
		}
		tm, err := time.Parse(time.RFC3339Nano, coin.LastUpdated)
		if err != nil {
			errMessage("Could not parse time string from API.", true)
		}
		tPrint("UPD:"+tm.Format(time.RFC822), list.lastUpdated, list, color.BgDarkGray, list.lastUpdatedWidth)
		if len(coin.Tickers) > 0 {
			tPrint(coin.Tickers[tickerIdx].Volume, list.volume, list, color.BgDarkGray, list.volumeWidth, "VOL:")
		} else {
			tPrint("no volume", list.volume, list, color.BgDarkGray, list.volumeWidth, "VOL:")
		}
		tPrint(coin.BlockTimeInMinutes, list.blockTIM, list, color.BgDarkGray, list.blockTIMWidth, "BT:")
	}
	fmt.Println()
}

// Helper function for printing tickers.
func tPrint(ifc interface{}, chk bool, lst listing, col color.Color, wid int, labl ...string) {
	if chk {
		switch ifc.(type) {
		case string:
			if lst.color {
				col.Print(cenTextInRange(ifc.(string), wid))
			} else {
				fmt.Print(cenTextInRange(ifc.(string), wid))
			}
		case float64:
			var id string
			if len(labl) > 0 {
				id = labl[0]
			}
			var form string
			switch id {
			case "VOL:":
				form = "%.4f"
			case "BT:":
				form = "%2.1f"
			default:
				form = "%f"
			}
			if lst.color {
				col.Print(cenTextInRange(id+fmt.Sprintf(form, ifc), wid))
			} else {
				fmt.Print(cenTextInRange(id+fmt.Sprintf(form, ifc), wid))
			}
		}
	}
}

// Performs an HTTP request.
func httpRequest(coinSymbol, userAgent string) (contents []byte, err error) {
	cli := http.Client{}
	req, err := http.NewRequest("GET", cgapi.CGCoinURL+coinSymbol, nil)
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
func cenTextInRange(str string, rng int) string {
	if utf8.RuneCountInString(str) > rng {
		if rng > 4 {
			str = str[:rng-5] + "."
		} else {
			str = "."
		}
	}
	diff := rng - utf8.RuneCountInString(str)
	buf := new(bytes.Buffer)
	if diff%2 != 0 {
		str = str + " "
	}
	for i := 0; i < diff/2; i++ {
		buf.WriteString(" ")
	}
	buf.WriteString(str)
	for i := 0; i < diff/2; i++ {
		buf.WriteString(" ")
	}
	return buf.String()
}

// Takes a map and returns an sorted slice of strings.
func mapToSortedStrings(mp map[string]string) []string {
	var keys = make([]string, len(mp))
	i := 0
	for k := range mp {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	return keys
}

// Displays all the available items in a string map exported from cgapi.
func listTableKeys(mp map[string]string, str string) {
	items := mapToSortedStrings(mp)
	fmt.Println("Available " + str + ":")
	for i, val := range items {
		fmt.Printf("%v\t%v\n", i, val)
	}
}

// Give user an error message and sometimes exit.
func errMessage(str string, exit bool, lst ...listing) {
	if len(lst) > 0 {
		tPrint("error", true, lst[0], color.BgRed, 9)
		tPrint(str, true, lst[0], color.FgDefault, len(str)+4)
		if exit {
			os.Exit(1)
		}
	} else {
		log.Fatal(str)
	}
	fmt.Println()
}
