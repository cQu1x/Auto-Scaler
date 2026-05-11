package handlers

type WorkRequest struct {
	Amount   int `json:"amount"`
	Duration int `json:"duration"`
}
