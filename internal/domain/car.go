package domain

type Car struct {
	ID             int64  `json:"id"`
	CarName        string `json:"car_name"`
	CarNumber      string `json:"car_number"`
	Status         string `json:"status"`
	PricePerMinute int64  `json:"price_per_minute"`
}
