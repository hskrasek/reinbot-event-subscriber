package main

type Payload struct {
	Type string `json:"type"`
	CallbackId string `json:"callback_id"`
	Channel Channel `json:"channel"`
	User User `json:"user"`
	Message Message `json:"message"`
	ActionTime string `json:"action_ts"`
	ResponseUrl string `json:"response_url"`
}

func (p Payload) isAction() bool {
	return p.Type == "message_action"
}