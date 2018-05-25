package main

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

func getBotByToken(token string) *Bot {
	var bot Bot
	orm.DB.First(&bot, "token = ?", token)

	return &bot
}

func (b *Bot) createBot() error {
	return orm.DB.Create(b).Error
}

func (b *Bot) setBotActivity() error {
	return orm.DB.Model(b).Where("token = ?", b.Token).Update("Active", !b.Active).Error
}

func getBotChannelByToken(token string) uint64 {
	var b Bot
	orm.DB.First(&b, "token = ?", token)

	return b.Channel
}

func (b *Bots) getBotsByClientID(uid string) error {
	return orm.DB.Where("client_id = ?", uid).Find(b).Error
}
