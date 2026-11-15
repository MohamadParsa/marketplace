package store

type Store struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Owner  string `json:"owner"`
	City   string `json:"city"`
	Active bool   `json:"active"`
	Model  string `json:"model"`
}
