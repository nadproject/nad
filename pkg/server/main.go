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

	"github.com/nadproject/nad/pkg/server/config"
	"github.com/nadproject/nad/pkg/server/models"
	"github.com/nadproject/nad/pkg/server/routes"
)

var versionTag = "master"
var templateDir = flag.String("templateDir", "tpl/web", "the path to a directory containing templates")

func newServer(c config.Config, s *models.Services) *http.ServeMux {
	//	apiRouter, err := c.NewAPI()
	//	if err != nil {
	//		panic(errors.Wrap(err, "initializing api router"))
	//	}

	webRouter := routes.NewWeb(c, s)

	mux := http.NewServeMux()
	// mux.Handle("/api/", http.StripPrefix("/api", apiRouter))
	mux.Handle("/", webRouter)

	return mux
}

func startCmd() {
	c := config.Load()

	services, err := models.NewServices(
		models.WithGorm("postgres", c.DB.GetConnectionStr()),
		models.WithUser(),
	)
	must(err)
	defer services.Close()

	err = services.AutoMigrate()
	must(err)

	srv := newServer(c, services)
	log.Printf("nad version %s is running on port %s", versionTag, c.Port)
	log.Fatalln(http.ListenAndServe(":"+c.Port, srv))
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

func must(err error) {
	if err != nil {
		panic(err)
	}
}
