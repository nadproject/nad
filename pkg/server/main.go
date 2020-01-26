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

	"github.com/nadproject/nad/pkg/clock"
	"github.com/nadproject/nad/pkg/server/buildinfo"
	"github.com/nadproject/nad/pkg/server/config"
	"github.com/nadproject/nad/pkg/server/controllers"
	"github.com/nadproject/nad/pkg/server/models"
	"github.com/nadproject/nad/pkg/server/routes"
)

var pageDir = flag.String("pageDir", "views", "the path to a directory containing page templates")
var staticDir = flag.String("staticDir", "./static/", "the path to the static directory ")

func startCmd() {
	cfg := config.Load()
	cfg.SetPageTemplateDir(*pageDir)
	cfg.SetStaticDir(*staticDir)

	services, err := models.NewServices(models.DialectPostgres, cfg.DB.GetConnectionStr())
	must(err)
	defer services.Close()

	err = services.SetupDB()
	must(err)

	cl := clock.New()
	ctl := controllers.New(cfg, services, cl)
	rc := routes.RouteConfig{
		WebRoutes:   routes.NewWebRoutes(cfg, ctl, services, cl),
		APIRoutes:   routes.NewAPIRoutes(cfg, ctl, services, cl),
		Controllers: ctl,
	}
	r := routes.New(cfg, services, rc)

	log.Printf("nad version %s is running on port %s", buildinfo.Version, cfg.Port)
	log.Fatalln(http.ListenAndServe(fmt.Sprintf(":%s", cfg.Port), r))
}

func versionCmd() {
	fmt.Printf("nad-server-%s\n", buildinfo.Version)
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
