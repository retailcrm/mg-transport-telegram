package main

import (
	"time"

	"github.com/getsentry/raven-go"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// Orm struct
type Orm struct {
	DB *gorm.DB
}

// NewDb init new database connection
func NewDb(config *TransportConfig) *Orm {
	db, err := gorm.Open("postgres", config.Database.Connection)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		panic(err)
	}

	db.DB().SetConnMaxLifetime(time.Duration(config.Database.ConnectionLifetime) * time.Second)
	db.DB().SetMaxOpenConns(config.Database.MaxOpenConnections)
	db.DB().SetMaxIdleConns(config.Database.MaxIdleConnections)

	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		return config.Database.TablePrefix + defaultTableName
	}

	db.SingularTable(true)
	db.LogMode(config.Database.Logging)

	setCreatedAt := func(scope *gorm.Scope) {
		if scope.HasColumn("CreatedAt") {
			scope.SetColumn("CreatedAt", time.Now())
		}
	}

	db.Callback().Create().Replace("gorm:update_time_stamp", setCreatedAt)

	return &Orm{
		DB: db,
	}
}

// Close connection
func (orm *Orm) Close() {
	orm.DB.Close()
}
