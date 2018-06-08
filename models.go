package main

import "time"

// Connection model
type Connection struct {
	ID        int    `gorm:"primary_key"`
	ClientID  string `gorm:"client_id type:varchar(70);not null;unique" json:"clientId,omitempty"`
	APIKEY    string `gorm:"api_key type:varchar(100);not null" json:"api_key,omitempty"`
	APIURL    string `gorm:"api_url type:varchar(255);not null" json:"api_url,omitempty"`
	MGURL     string `gorm:"mg_url type:varchar(255);not null;" json:"mg_url,omitempty"`
	MGToken   string `gorm:"mg_token type:varchar(100);not null;unique" json:"mg_token,omitempty"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Active    bool  `json:"active,omitempty"`
	Bots      []Bot `gorm:"foreignkey:ConnectionID"`
}

// Bot model
type Bot struct {
	ID           int    `gorm:"primary_key"`
	ConnectionID int    `gorm:"connection_id" json:"connectionId,omitempty"`
	Channel      uint64 `gorm:"channel;not null;unique" json:"channel,omitempty"`
	Token        string `gorm:"token type:varchar(100);not null;unique" json:"token,omitempty"`
	Name         string `gorm:"name type:varchar(40)" json:"name,omitempty"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Active       bool `json:"active,omitempty"`
}

// Users model
type Users struct {
	ID           int    `gorm:"primary_key"`
	ExternalID   int    `gorm:"external_id;not null;unique"`
	UserPhotoURL string `gorm:"user_photo_url type:varchar(255);unique"`
	UserPhotoID  string `gorm:"user_photo_id type:varchar(100);unique"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

//Bots list
type Bots []Bot
