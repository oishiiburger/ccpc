// cgapi.go
// Some structs and hash maps for the Coin Gecko API are herein.
// This is not exhaustive, and probably never will be.package CGapi

package cgapi

const (
	// CGCoinURL is the API URL for the upmost coin array.
	CGCoinURL string = "https://api.coingecko.com/api/v3/coins"
)

// CGCoinURLs is a mapping of coin symbols to API URLs.
var CGCoinURLs = map[string]string{
	"btc": "https://api.coingecko.com/api/v3/coins/bitcoin",
	"eth": "https://api.coingecko.com/api/v3/coins/ethereum",
	"xrp": "https://api.coingecko.com/api/v3/coins/ripple",
	//	"usdt": "https://api.coingecko.com/api/v3/coins/tether",
	"bch": "https://api.coingecko.com/api/v3/coins/bitcoin-cash",
	"ltc": "https://api.coingecko.com/api/v3/coins/litecoin",
	"eos": "https://api.coingecko.com/api/v3/coins/eos",
	"bnb": "https://api.coingecko.com/api/v3/coins/binancecoin",
	"bsv": "https://api.coingecko.com/api/v3/coins/bitcoin-cash-sv"}

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
