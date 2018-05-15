package main

// SiteBot methods
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
	return orm.DB.Model(c).Where("client_id = ?", c.ClientID).Update("Active", c.Active).Error
}

func (c *Connection) create() error {
	return orm.DB.Create(c).Error
}

func (c *Connection) save() error {
	return orm.DB.Model(c).Where("client_id = ?", c.ClientID).Update(c).Error
}

// Bot methods
func getByToken(token string) (*Bot, error) {
	var bot Bot
	orm.DB.First(&bot, "token = ?", token)

	return &bot, nil
}

func (b *Bot) createBot() error {
	return orm.DB.Create(b).Error
}

func (b *Bot) deleteBot() error {
	return orm.DB.Where("token = ?", b.Token).Delete(b).Error
}

func (b *Bots) getBotsByClientId(uid string) error {
	return orm.DB.Where("client_id = ?", uid).Find(b).Error
}
