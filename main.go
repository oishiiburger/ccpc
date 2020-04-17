// ccpc crypto coin price checker

package main

import (
	"bytes"
	"encoding/json"
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
	userAgent string = "ccpc"
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
	self.nameWidth = 16
	self.priceWidth = 14
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
	allPtr := flag.BoolP("all", "a", false, "Yields listings for all supported coins. (Generally not recommended)")
	blkPtr := flag.BoolP("block-time", "b", false, "Includes block time in the listing, if available.")
	bwtPtr := flag.BoolP("no-color", "c", false, "Disables output colors.")
	maxPtr := flag.BoolP("maximum", "m", false, "Yields maximum detail listings for the selected coins.")
	namPtr := flag.BoolP("no-name", "n", false, "Omits coin name in the listing.")
	tgtPtr := flag.StringP("target", "t", "usd", "Determines the target currency for comparison (e.g. usd, jpy).")
	timPtr := flag.BoolP("no-time", "z", false, "Omits last update time in the listing.")
	volPtr := flag.BoolP("volume", "v", false, "Includes coin volume in the listing, if available.")
	lcPtr := flag.Bool("list-coins", false, "Displays a listing of all supported coins.")
	lmPtr := flag.Bool("list-currencies", false, "Displays a listing of all supported currencies.")
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
			errMessage("Unsupported target currency: "+tgt, listingProps)
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
			if !strings.Contains(arg, "-") && len(cgapi.CGCoinURLs[arg]) > 0 {
				args = append(args, arg)
			}
		}
		for _, arg := range args {
			res, _ := httpRequest(cgapi.CGCoinURLs[arg], userAgent)
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
		errMessage("Symbol not found.")
	} else {
		tPrint(coin.Symbol, list.symbol, list, color.BgBlue, list.symbolWidth)
		tPrint(coin.Name, list.name, list, color.FgBlue, list.nameWidth)
		if len(coin.Tickers) > 0 { // check to make sure this coin has a ticker
			for tickerIdx, t := range coin.Tickers {
				if t.Target == list.target {
					if coin.MarketData.PriceChange24h >= 0 {
						tPrint(cgapi.MonetarySymbols[list.target]+fmt.Sprintf("%.2f", coin.Tickers[tickerIdx].Last),
							true, list, color.BgGreen, list.priceWidth)
					} else {
						tPrint(cgapi.MonetarySymbols[list.target]+fmt.Sprintf("%.2f", coin.Tickers[tickerIdx].Last),
							true, list, color.BgRed, list.priceWidth)
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
			errMessage("Could not parse time string from API.")
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
func tPrint(ifc interface{}, chk bool, lst listing, col color.Color, wid int, str ...string) {
	switch ifc.(type) {
	case string:
		if lst.color {
			col.Print(cenTextInRange(ifc.(string), wid))
		} else {
			fmt.Print(cenTextInRange(ifc.(string), wid))
		}
	case float64:
		var id string
		if len(str) > 0 {
			id = str[0]
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
		errMessage("A requested string was too long to fit within a column.")
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

// Give user an irrecoverable error message and exit.
func errMessage(str string, lst ...listing) {
	if len(lst) > 0 {
		if lst[0].color {
			color.BgRed.Print("  error  ")
		} else {
			fmt.Print("  error  ")
		}
	} else {
		fmt.Print("  error  ")
	}
	fmt.Println(" " + str)
	os.Exit(1)
}
