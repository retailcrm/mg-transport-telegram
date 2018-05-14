package main

import "time"

// Connection model
type Connection struct {
	ID        int    `gorm:"primary_key"`
	ClientID  string `json:"clientId,omitempty"`
	APIKEY    string `json:"api_key,omitempty"`
	APIURL    string `gorm:"url_crm" json:"url,omitempty"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Number    string `json:"number,omitempty"`
	Active    bool   `gorm:"active" json:"active,omitempty"`
}

// Bot model
type Bot struct {
	ID        int    `gorm:"primary_key"`
	ClientID  string `json:"clientId,omitempty"`
	Token     string `json:"token,omitempty"`
	Name      string `json:"name,omitempty"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Active    bool `gorm:"active" json:"active,omitempty"`
}

type SiteBot struct {
	ID       int    `gorm:"primary_key"`
	SiteCode string `json:"siteCode,omitempty"`
	BotId    string `json:"botId,omitempty"`
}

//Bots list
type Bots []Bot

func createSiteBots(s []SiteBot) error {
	tx := orm.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, val := range s {
		if err := tx.Create(&val).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// Connection methods
func get(uid string) (*Connection, error) {
	var connection Connection
	orm.DB.First(&connection, "client_id = ?", uid)

	return &connection, nil
}

func getByUrlCrm(urlCrm string) (*Connection, error) {
	var connection Connection
	orm.DB.First(&connection, "api_url = ?", urlCrm)

	return &connection, nil
}

func (c *Connection) setActive() error {
	return orm.DB.Model(&c).Where("client_id = ?", c.ClientID).Update("Active", c.Active).Error
}

func (c *Connection) create() error {
	return orm.DB.Create(&c).Error
}

func (c *Connection) save() error {
	return orm.DB.Model(&c).Where("client_id = ?", c.ClientID).Update(c).Error
}

// Bot methods
func getByToken(token string) (*Bot, error) {
	var bot Bot
	orm.DB.First(&bot, "token = ?", token)

	return &bot, nil
}

func (b *Bot) createBot() error {
	return orm.DB.Create(&b).Error
}

func (b *Bot) deleteBot() error {
	return orm.DB.Where("token = ?", b.Token).Delete(&b).Error
}

func (b *Bots) getBotsByClientId(uid string) error {
	return orm.DB.Where("client_id = ?", uid).Find(&b).Error
}
