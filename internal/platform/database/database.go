package database

import (
	"fmt"

	"github.com/sebaactis/wallet-go-api/internal/platform/config"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Open(cfg config.Config) (*gorm.DB, error) {

	switch cfg.Driver {
	case "sqlite":
		return gorm.Open(sqlite.Open(cfg.DSN), &gorm.Config{})
	default:
		return nil, fmt.Errorf("Driver not supported: %s", cfg.Driver)
	}
}

func Migrate(db *gorm.DB) error {

	// TO DO ---> Migraciones de modelos etc
	return nil
}
