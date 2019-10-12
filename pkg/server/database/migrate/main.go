/* Copyright (C) 2019 Monomax Software Pty Ltd
 *
 * This file is part of NAD.
 *
 * NAD is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * NAD is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with NAD.  If not, see <https://www.gnu.org/licenses/>.
 */

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/nadproject/nad/pkg/server/database"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"github.com/rubenv/sql-migrate"
)

var (
	migrationDir = flag.String("migrationDir", "../migrations", "the path to the directory with migraiton files")
)

func init() {
	fmt.Println("Migrating NAD database...")

	// Load env
	if os.Getenv("GO_ENV") != "PRODUCTION" {
		if err := godotenv.Load("../../.env.dev"); err != nil {
			panic(err)
		}
	}

	c := database.Config{
		Host:     os.Getenv("DBHost"),
		Port:     os.Getenv("DBPort"),
		Name:     os.Getenv("DBName"),
		User:     os.Getenv("DBUser"),
		Password: os.Getenv("DBPassword"),
	}
	database.Open(c)
}

func main() {
	flag.Parse()

	db := database.DBConn

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
