package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/nadproject/nad/pkg/server/dbconn"
	"github.com/pkg/errors"
	"github.com/rubenv/sql-migrate"
)

var (
	migrationDir = flag.String("migrationDir", "../sql", "the path to the directory with migraiton files")
)

func init() {
	fmt.Println("Migrating nad database...")

	// Load env
	if os.Getenv("GO_ENV") != "PRODUCTION" {
		if err := godotenv.Load("../../.env.dev"); err != nil {
			panic(err)
		}
	}

}

func main() {
	flag.Parse()

	db := dbconn.Open(dbconn.Config{
		Host:     os.Getenv("DBHost"),
		Port:     os.Getenv("DBPort"),
		Name:     os.Getenv("DBName"),
		User:     os.Getenv("DBUser"),
		Password: os.Getenv("DBPassword"),
	})

	migrations := &migrate.FileMigrationSource{
		Dir: *migrationDir,
	}

	migrate.SetTable("migrations")

	n, err := migrate.Exec(db.DB(), "postgres", migrations, migrate.Up)
	if err != nil {
		panic(errors.Wrap(err, "executing migrations"))
	}

	fmt.Printf("Applied %d migrations\n", n)
}
