package database

import (
	"fmt"

	"github.com/sebaactis/wallet-go-api/internal/account"
	ledger "github.com/sebaactis/wallet-go-api/internal/legder"
	"github.com/sebaactis/wallet-go-api/internal/platform/config"
	"github.com/sebaactis/wallet-go-api/internal/transaction"
	"github.com/sebaactis/wallet-go-api/internal/user"

	// sqlite original de GORM (COMÉNTALO si pasas a pure Go)
	// "gorm.io/driver/sqlite"

	// Driver “pure Go” basado en modernc (no requiere CGO)
	sqlite "github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func Open(cfg config.Config) (*gorm.DB, error) {
	switch cfg.Driver {
	case "sqlite":
		// Ejemplos de DSN válidos:
		// "wallet.db"
		// "file:wallet.db?_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)"
		return gorm.Open(sqlite.Open(cfg.DSN), &gorm.Config{})
	default:
		return nil, fmt.Errorf("Driver not supported: %s", cfg.Driver)
	}
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&user.User{},
		&account.Account{},
		&transaction.Transaction{},
		&ledger.LedgerEntry{},
	)
}
