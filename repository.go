package main

import "github.com/jinzhu/gorm"

func getConnection(uid string) *Connection {
	var connection Connection
	orm.DB.First(&connection, "client_id = ?", uid)

	return &connection
}

func getConnectionByURL(urlCrm string) *Connection {
	var connection Connection
	orm.DB.First(&connection, "api_url = ?", urlCrm)

	return &connection
}

func (c *Connection) setConnectionActivity() error {
	return orm.DB.Model(c).Where("client_id = ?", c.ClientID).Update("Active", c.Active).Error
}

func (c *Connection) createConnection() error {
	return orm.DB.Create(c).Error
}

func (c *Connection) saveConnection() error {
	return orm.DB.Model(c).Where("client_id = ?", c.ClientID).Update(c).Error
}

func (c *Connection) createBot(b Bot) error {
	return orm.DB.Model(c).Association("Bots").Append(&b).Error
}

func getConnectionByBotToken(token string) (*Connection, error) {
	var c Connection
	err := orm.DB.Where("active = ?", true).
		Preload("Bots", "token = ?", token).
		First(&c).Error
	if gorm.IsRecordNotFoundError(err) {
		return &c, nil
	} else {
		return &c, err
	}

	return &c, nil
}

func getBotByChannel(ch uint64) *Bot {
	var bot Bot
	orm.DB.First(&bot, "channel = ?", ch)

	return &bot
}

func (b *Bot) setBotActivity() error {
	return orm.DB.Model(b).Where("token = ?", b.Token).Update("Active", !b.Active).Error
}

func getBotChannelByToken(token string) uint64 {
	var b Bot
	orm.DB.First(&b, "token = ?", token)

	return b.Channel
}

func (c Connection) getBotsByClientID() Bots {
	var b Bots
	err := orm.DB.Model(c).Association("Bots").Find(&b).Error
	if err != nil {
		logger.Error(err)
	}

	return b
}

func getConnectionById(id int) *Connection {
	var connection Connection
	orm.DB.First(&connection, "id = ?", id)

	return &connection
}
