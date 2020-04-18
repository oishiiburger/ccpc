// cgapi.go
// Some structs and hash tables for the Coin Gecko API are here.
// This is not exhaustive, and probably never will be.

package cgapi

// APIPingURL is the URL for the server OK message.
const APIPingURL string = "https://api.coingecko.com/api/v3/ping"

// APIPing is a struct for the server OK message.
type APIPing struct {
	PingMsg string `json:"gecko_says"`
}

// MonetarySymbols is a mapping of currency abbreviations to symbols.
var MonetarySymbols = map[string]string{
	"BTC": "btc",
	"ETH": "eth",
	"LTC": "ltc",
	"BCH": "bch",
	"BNB": "bnb",
	"EOS": "eos",
	"XRP": "xrp",
	"XLM": "xlm",
	"USD": "$",
	"AED": "AED",
	"ARS": "ARS$",
	"AUD": "AUS$",
	"BDT": "৳",
	"BHD": "BHD",
	"BMD": "BMD",
	"BRL": "R$",
	"CAD": "CAD$",
	"CHF": "CHF",
	"CLP": "CLP$",
	"CNY": "元",
	"CZK": "Kč",
	"DKK": "kr",
	"EUR": "€",
	"GBP": "£",
	"HKD": "HK$",
	"HUF": "ft",
	"IDR": "Rp",
	"ILS": "₪",
	"INR": "₹",
	"JPY": "¥",
	"KRW": "₩",
	"KWD": "KWD",
	"LKR": "Rs",
	"MMK": "MMK",
	"MXN": "MXN$",
	"MYR": "RM",
	"NOK": "kr",
	"NZD": "NZD$",
	"PHP": "₱",
	"PKR": "Rs",
	"PLN": "zł",
	"RUB": "₽",
	"SAR": "SAR",
	"SEK": "kr",
	"SGD": "SGD$",
	"THB": "฿",
	"TRY": "₺",
	"TWD": "TWD",
	"UAH": "₴",
	"VEF": "Bs",
	"VND": "₫",
	"ZAR": "R$",
	"XDR": "XDR",
	"XAG": "XAG",
	"XAU": "XAU"}

// MonetaryNames is a mapping of currency abbreviations to names.
var MonetaryNames = map[string]string{
	"BTC": "Bitcoin",
	"ETH": "Etherium",
	"LTC": "Litecoin",
	"BCH": "Bitcoin Cash",
	"BNB": "Binance Coin",
	"EOS": "EOS",
	"XRP": "XRP",
	"XLM": "Stellar Lumens",
	"USD": "United States Dollar",
	"AED": "United Arab Emirates Dirham",
	"ARS": "Argentine Peso",
	"AUD": "Australian Dollar",
	"BDT": "Bangladeshi Taka",
	"BHD": "Bahrain Dinar",
	"BMD": "Bermuda Dollar",
	"BRL": "Brazilian Real",
	"CAD": "Canadian Dollar",
	"CHF": "Swiss Franc",
	"CLP": "Chilean Peso",
	"CNY": "Chinese Yuan",
	"CZK": "Czech Koruna",
	"DKK": "Danish Krone",
	"EUR": "Euro",
	"GBP": "United Kingdom Pound Sterling",
	"HKD": "Hong Kong Dollar",
	"HUF": "Hungarian Forint",
	"IDR": "Indonesian Rupiah",
	"ILS": "Israeli Shekel",
	"INR": "Indian Rupee",
	"JPY": "Japanese Yen",
	"KRW": "South Korean Won",
	"KWD": "Kuwait Dinar",
	"LKR": "Sri Lankan Rupee",
	"MMK": "Myanmar Kyat",
	"MXN": "Mexican Peso",
	"MYR": "Malaysian Ringgit",
	"NOK": "Norwegian Krone",
	"NZD": "New Zealand Dollar",
	"PHP": "Philippine Peso",
	"PKR": "Pakistan Rupee",
	"PLN": "Polish Złoty",
	"RUB": "Russian Ruble",
	"SAR": "Saudi Arabian Riyal",
	"SEK": "Swedish Krona",
	"SGD": "Singapore Dollar",
	"THB": "Thai Baht",
	"TRY": "Turkish Lira",
	"TWD": "Taiwan Dollar",
	"UAH": "Ukrainian Hryvna",
	"VEF": "Venezuelan Bolívar",
	"VND": "Vietnamese Dong",
	"ZAR": "South African Rand",
	"XDR": "IMF Special Drawing Rights",
	"XAG": "Silver Ounce",
	"XAU": "Gold Ounce"}

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
