package main

type EventMessage struct {
	ChannelID  uint64 `json:"channel_id"`
	ExternalID int    `json:"external_id,omitempty"`
	ChatID     int64  `json:"chat_id"`
	Message    string `json:"message,omitempty"`
	Type       string `json:"type"`
}
