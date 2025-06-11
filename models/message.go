package models

type Message struct {
	Message string `json:"message"`
	From string `json:"from"`
	To string `json:"to"`
}