package models

type Asset struct {
	ID        string  `json:"id"`
	Symbol    string  `json:"symbol"`
	VolumeUSD float64 `json:"volume_usd,omitempty"`
}

type TokenDay struct {
	TokenID   string  `json:"token_id"`
	VolumeUSD float64 `json:"volume_usd`
	Date      int64   `json:"date"`
}

type VolumePerTimeWindow struct {
	TokenId        string  `json:token_id`
	StartTime      int64   `json:"start_time,string"`
	EndTime        int64   `json:"end_time,string"`
	TotalVolumeUSD float64 `json:"total_volume_USD"`
}

type Pool struct {
	ID           string `json:"id"`
	Asset0Symbol string `json:"asset0_symbol"`
	Asset1Symbol string `json:"asset1_symbol"`
}

type Transaction struct {
	ID          string `json:"id"`
	BlockNumber int64  `json:"block_number"`
	Swaps       []Swap `json:"swaps"`
}

type Swap struct {
	ID      string  `json:"id"`
	Amount0 float64 `json:amount0`
	Amount1 float64 `json:amount1`
	Asset0  Asset   `json:"asset0"`
	Asset1  Asset   `json:"asset1"`
}

type SwapResult struct {
	BlockNumber int64   `json:"block_number"`
	Swaps       []Swap  `json:"swaps"`
	Assets      []Asset `json:"assets"`
}
