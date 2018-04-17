package helper

import ("encoding/json")

type Response struct {
	Status string `json:"status"`
	Message string `json:"message"`
	Data json.RawMessage `json:"data"`
}