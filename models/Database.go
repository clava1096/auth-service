package models

import (
	"auth-service/connections"
	"github.com/gofiber/fiber/v2/log"
)

func Migrate() {
	log.Info("Running migrations database")
	migrate := connections.DB.AutoMigrate(
		&Token{},
		&User{},
	)
	if migrate != nil {
		log.Panicf("Failed to migrate database: %s", migrate)
	}

}
