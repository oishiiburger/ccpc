// ccpc crypto coin price checker

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"

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
	"EUR": "€"}

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
		// for k, v := CGCoinURLs {

		// }
	}

	res, _ := httpRequest(CGCoinURLs["usdt"], userAgent)
	var coin CGCoinSingleton
	json.Unmarshal(res, &coin)

	var testlisting listing
	testlisting.color = false
	testlisting.name = true
	testlisting.target = "GBP"
	displayCoinListing(coin, defaultListing())
}

// Display coin listing
func displayCoinListing(coin CGCoinSingleton, list listing) {
	buf := new(bytes.Buffer) // buffer is only used when not in color mode

	if list.color {
		color.New(color.FgBlack, color.BgBlue).Print(" " + coin.Symbol + " ")
		fmt.Print(" ")
	} else {
		buf.WriteString(" " + coin.Symbol + "  ")
	}

	if list.name {
		if list.color {
			color.FgBlue.Print(coin.Name + " ")
		} else {
			buf.WriteString(fmt.Sprint(coin.Name + " "))
		}
	}

	for i, t := range coin.Tickers {
		if t.Target == list.target {
			if list.color {
				if coin.MarketData.PriceChange24h >= 0 {
					color.BgGreen.Print(" " + MonetarySymbols[list.target] +
						fmt.Sprintf("%.2f", coin.Tickers[i].Last) + " ")
				} else {
					color.BgRed.Print(" " + MonetarySymbols[list.target] +
						fmt.Sprintf("%.2f", coin.Tickers[i].Last) + " ")
				}
			} else {
				buf.WriteString(" " + MonetarySymbols[list.target] +
					fmt.Sprintf("%.2f", coin.Tickers[i].Last) + " ")
			}
			fmt.Print(" ")
			if list.volume {
				buf.WriteString(fmt.Sprintf("%.8f", coin.Tickers[i].Volume))
			}
			break
		}
		if i == len(coin.Tickers)-1 {
			buf.WriteString("N/A ")
		}
	}

	if list.lastUpdated || list.blockTimeInMinutes {
		buf.WriteString("\n")
		if list.lastUpdated {
			buf.WriteString("last updated: " + coin.LastUpdated)
			if list.blockTimeInMinutes {
				buf.WriteString(", ")
			} else {
				buf.WriteString(" ")
			}
		}
		if list.blockTimeInMinutes {
			buf.WriteString("block time: " + fmt.Sprintf("%.2f", coin.BlockTimeInMinutes) + " minutes")
		}
	}

	fmt.Println(buf.String())
}

// Performs an HTTP request
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
