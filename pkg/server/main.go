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
	"log"
	"net/http"
	"os"

	"github.com/jinzhu/gorm"
	"github.com/nadproject/nad/pkg/clock"
	"github.com/nadproject/nad/pkg/server/api"
	"github.com/nadproject/nad/pkg/server/app"
	"github.com/nadproject/nad/pkg/server/database"
	"github.com/nadproject/nad/pkg/server/dbconn"
	"github.com/nadproject/nad/pkg/server/mailer"

	"github.com/gobuffalo/packr/v2"
	"github.com/pkg/errors"
)

var versionTag = "master"
var port = flag.String("port", "3000", "port to connect to")
var rootBox *packr.Box

func init() {
	rootBox = packr.New("root", "../../web/public")
}

func initServer(a app.App) (*http.ServeMux, error) {
	c := api.Context{App: &a}

	apiRouter, err := c.New()
	if err != nil {
		return nil, errors.Wrap(err, "initializing api router")
	}

	mux := http.NewServeMux()
	mux.Handle("/api/", http.StripPrefix("/api", apiRouter))
	// mux.Handle("/", webRouter)

	return mux, nil
}

func initDB() *gorm.DB {
	db := dbconn.Open(dbconn.Config{
		Host:     os.Getenv("DBHost"),
		Port:     os.Getenv("DBPort"),
		Name:     os.Getenv("DBName"),
		User:     os.Getenv("DBUser"),
		Password: os.Getenv("DBPassword"),
	})
	database.InitSchema(db)

	return db
}

func initApp() app.App {
	db := initDB()

	return app.App{
		DB:               db,
		Clock:            clock.New(),
		StripeAPIBackend: nil,
		EmailTemplates:   mailer.NewTemplates(nil),
		EmailBackend:     &mailer.SimpleBackendImplementation{},
		Config: app.Config{
			WebURL:              os.Getenv("WebURL"),
			OnPremise:           true,
			DisableRegistration: os.Getenv("DisableRegistration") == "true",
		},
	}
}

func startCmd() {
	app := initApp()
	defer app.DB.Close()

	if err := database.Migrate(app.DB); err != nil {
		panic(errors.Wrap(err, "running migrations"))
	}

	srv, err := initServer(app)
	if err != nil {
		panic(errors.Wrap(err, "initializing server"))
	}

	log.Printf("nad version %s is running on port %s", versionTag, *port)
	log.Fatalln(http.ListenAndServe(":"+*port, srv))
}

func versionCmd() {
	fmt.Printf("nad-server-%s\n", versionTag)
}

func rootCmd() {
	fmt.Printf(`nad server - A simple notebook for developers

Usage:
  nad-server [command]

Available commands:
  start: Start the server
  version: Print the version
`)
}

func main() {
	flag.Parse()
	cmd := flag.Arg(0)

	switch cmd {
	case "":
		rootCmd()
	case "start":
		startCmd()
	case "version":
		versionCmd()
	default:
		fmt.Printf("Unknown command %s", cmd)
	}
}
