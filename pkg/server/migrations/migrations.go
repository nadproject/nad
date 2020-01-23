package migrations

import (
	"log"

	"github.com/gobuffalo/packr/v2"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/rubenv/sql-migrate"
)

var (
	// MigrationTableName is the name of the table that keeps track of migrations
	MigrationTableName = "migrations"
)

// Run runs the migrations
func Run(db *gorm.DB) error {
	migrations := &migrate.PackrMigrationSource{
		Box: packr.New("migrations", "./sql"),
	}

	migrate.SetTable(MigrationTableName)

	n, err := migrate.Exec(db.DB(), "postgres", migrations, migrate.Up)
	if err != nil {
		return errors.Wrap(err, "running migrations")
	}

	log.Printf("Performed %d migrations", n)

	return nil
}
