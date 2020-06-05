package models

// CallPut for unmarshalling API data
type CallPut struct {
	StrikePrice           float64 `json:"strikePrice"`
	ExpiryDate            string  `json:"expiryDate"`
	Underlying            string  `json:"underlying"`
	Identifier            string  `json:"identifier"`
	OpenInterest          float64 `json:"openInterest"`
	ChangeinOpenInterest  float64 `json:"changeinOpenInterest"`
	PchangeinOpenInterest float64 `json:"pchangeinOpenInterest"`
	TotalTradedVolume     int     `json:"totalTradedVolume"`
	ImpliedVolatility     float64 `json:"impliedVolatility"`
	LastPrice             float64 `json:"lastPrice"`
	Change                float64 `json:"change"`
	PChange               float64 `json:"pChange"`
	TotalBuyQuantity      int     `json:"totalBuyQuantity"`
	TotalSellQuantity     int     `json:"totalSellQuantity"`
	BidQty                float64 `json:"bidQty"`
	Bidprice              float64 `json:"bidprice"`
	AskQty                int     `json:"askQty"`
	AskPrice              float64 `json:"askPrice"`
	UnderlyingValue       float64 `json:"underlyingValue"`
}

//Share is struct
type Share struct {
	StrikePrice float64 `json:"strikePrice"`
	ExpiryDate  string  `json:"expiryDate"`
	PE          CallPut `json:"PE"`
	CE          CallPut `json:"CE"`
}
