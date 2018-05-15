package main

import "time"

// Connection model
type Connection struct {
	ID        int    `gorm:"primary_key"`
	ClientID  string `gorm:"client_id" json:"clientId,omitempty"`
	APIKEY    string `gorm:"api_key" json:"api_key,omitempty"`
	APIURL    string `gorm:"api_url" json:"url,omitempty"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Active    bool `json:"active,omitempty"`
}

// Bot model
type Bot struct {
	ID        int    `gorm:"primary_key"`
	ClientID  string `gorm:"client_id" json:"clientId,omitempty"`
	Token     string `json:"token,omitempty"`
	Name      string `json:"name,omitempty"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Active    bool `json:"active,omitempty"`
}

// SiteBot model
type SiteBot struct {
	ID       int    `gorm:"primary_key"`
	SiteCode string `gorm:"site_code" json:"siteCode,omitempty"`
	BotID    string `gorm:"bot_id" json:"botId,omitempty"`
}

//Bots list
type Bots []Bot
