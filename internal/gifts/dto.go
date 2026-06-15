package gifts

type CreateGiftRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	URL         string  `json:"url"`
	Reserved    bool    `json:"reserved"`
	ReservedBy  string  `json:"reserved_by"`
}

type UpdateGiftRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	URL         string  `json:"url"`
	Reserved    bool    `json:"reserved"`
	ReservedBy  string  `json:"reserved_by"`
}
