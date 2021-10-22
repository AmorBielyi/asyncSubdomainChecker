package models

// Struct that represents JSON Response
type Result struct {
	Subdomain string `json:"subdomain"`
	Status    string `json:"status"`
	Code      int    `json:"code"`
}
