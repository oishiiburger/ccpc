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

	"github.com/gookit/color"
)

const (
	userAgent string = "ccpc"

	// CGCoinURL is the API URL for the upmost coin array.
	CGCoinURL string = "https://api.coingecko.com/api/v3/coins"
)

// CGCoinURLs is a mapping of coin symbols to API URLs.
var CGCoinURLs = map[string]string{
	"btc":  "https://api.coingecko.com/api/v3/coins/bitcoin",
	"eth":  "https://api.coingecko.com/api/v3/coins/ethereum",
	"xrp":  "https://api.coingecko.com/api/v3/coins/ripple",
	"usdt": "https://api.coingecko.com/api/v3/coins/tether",
	"bch":  "https://api.coingecko.com/api/v3/coins/bitcoin-cash",
	"ltc":  "https://api.coingecko.com/api/v3/coins/litecoin",
	"eos":  "https://api.coingecko.com/api/v3/coins/eos",
	"bnb":  "https://api.coingecko.com/api/v3/coins/binancecoin",
	"bsv":  "https://api.coingecko.com/api/v3/coins/bitcoin-cash-sv"}

// MonetarySymbols is a mapping of currency abbreviations to symbols.
var MonetarySymbols = map[string]string{
	"USD": "$",
	"GBP": "£",
	"JPY": "¥",
	"EUR": "€",
	"BTC": "btc"}

// CGCoin defines a coin and its features.
type CGCoin struct {
	ID                 string `json:"id"`
	Symbol             string `json:"symbol"`
	Name               string `json:"name"`
	BlockTimeInMinutes string `json:"block_time_in_minutes"`
	LastUpdated        string `json:"last_updated"`
}

// CGCoinSingleton defines a singleton coin and its features.
type CGCoinSingleton struct {
	ID                 string           `json:"id"`
	Symbol             string           `json:"symbol"`
	Name               string           `json:"name"`
	BlockTimeInMinutes float64          `json:"block_time_in_minutes"`
	LastUpdated        string           `json:"last_updated"`
	Tickers            []CGTicker       `json:"tickers"`
	MarketData         CGCoinMarketData `json:"market_data"`
}

// CGCoinMarketData encapsulates price change data over time.
type CGCoinMarketData struct {
	PriceChange24h   float64 `json:"price_change_24h"`
	PriceChange24hPc float64 `json:"price_change_percentage_24h"`
	// others exist in the JSON
}

// CGTicker defines tickers which exist for each coin.
type CGTicker struct {
	Base       string  `json:"base"`
	Target     string  `json:"target"`
	Last       float64 `json:"last"`
	Volume     float64 `json:"volume"`
	TrustScore string  `json:"trust_score"`
	Timestamp  string  `json:"timestamp"`
}

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

func main() {
	// CLI flag handling
	allPtr := flag.Bool("all", false, "Yields default listings for all supported coins.")
	flag.Parse()

	// -all
	if *allPtr {
		var keys = make([]string, len(CGCoinURLs))
		i := 0
		for k := range CGCoinURLs {
			keys[i] = k
			i++
		}
		sort.Strings(keys)
		for key := 0; key < len(keys); key++ {
			res, _ := httpRequest(CGCoinURLs[keys[key]], userAgent)
			var coin CGCoinSingleton
			json.Unmarshal(res, &coin)
			displayCoinListing(coin, defaultListing(), true)
		}
	}

	// res, _ := httpRequest(CGCoinURLs["btc"], userAgent)
	// var coin CGCoinSingleton
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
func displayCoinListing(coin CGCoinSingleton, list listing, leftAlign bool) {
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
						color.BgGreen.Print(ctir(MonetarySymbols[list.target]+
							fmt.Sprintf("%.2f", coin.Tickers[i].Last), 14, leftAlign))
					} else {
						color.BgRed.Print(ctir(MonetarySymbols[list.target]+
							fmt.Sprintf("%.2f", coin.Tickers[i].Last), 14, leftAlign))
					}
				} else {
					buf.WriteString(ctir(MonetarySymbols[list.target]+
						fmt.Sprintf("%.2f", coin.Tickers[i].Last), 14, leftAlign))
				}
				if list.volume {
					if list.color {
						color.BgDarkGray.Print(ctir("VOL:"+fmt.Sprintf("%.8f", coin.Tickers[i].Volume), 20))
					} else {
						buf.WriteString(ctir("VOL:"+fmt.Sprintf("%.8f", coin.Tickers[i].Volume), 20))
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
					color.BgDarkGray.Print(ctir("BT:"+fmt.Sprintf("%.2f", coin.BlockTimeInMinutes)+" MINS", 15))
				} else {
					buf.WriteString(ctir("BT:"+fmt.Sprintf("%.2f", coin.BlockTimeInMinutes)+" MINS", 15))
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
			buf.WriteString(str + " ")
			for i := 0; i < (diff-1)/2; i++ {
				buf.WriteString(" ")
			}
		}
	}
	return buf.String()
}
