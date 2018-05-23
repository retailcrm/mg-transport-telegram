package main

import "time"

// Connection model
type Connection struct {
	ID        int    `gorm:"primary_key"`
	ClientID  string `gorm:"client_id" json:"clientId,omitempty"`
	APIKEY    string `gorm:"api_key" json:"api_key,omitempty"`
	APIURL    string `gorm:"api_url" json:"api_url,omitempty"`
	MGURL     string `gorm:"mg_url" json:"mg_url,omitempty"`
	MGToken   string `gorm:"mg_token" json:"mg_token,omitempty"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Active    bool `json:"active,omitempty"`
}

// Bot model
type Bot struct {
	ID        int    `gorm:"primary_key"`
	ClientID  string `gorm:"client_id" json:"clientId,omitempty"`
	Channel   uint64 `json:"channel,omitempty"`
	Token     string `json:"token,omitempty"`
	Name      string `json:"name,omitempty"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Active    bool `json:"active,omitempty"`
}

//Bots list
type Bots []Bot
