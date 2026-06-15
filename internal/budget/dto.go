package budget

type CreateItemRequest struct {
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Amount      float64 `json:"amount"`
	Paid        bool    `json:"paid"`
}

type UpdateItemRequest struct {
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Amount      float64 `json:"amount"`
	Paid        bool    `json:"paid"`
}
