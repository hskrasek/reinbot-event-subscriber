package main

type Message struct {
	Type string `json:"type"`
	UserId string `json:"user"`
	Ts string `json:"ts"`
	Text string `json:"text"`
}