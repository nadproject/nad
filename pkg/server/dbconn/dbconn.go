/* Copyright (C) 2019 Monomax Software Pty Ltd
 *
 * This file is part of nad.
 *
 * nad is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * nad is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with nad.  If not, see <https://www.gnu.org/licenses/>.
 */

package dbconn

import (
	"fmt"
	"os"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

// Config holds the connection configuration
type Config struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
}

// ErrConfigMissingHost is an error for an incomplete configuration missing the host
var ErrConfigMissingHost = errors.New("Host is empty")

// ErrConfigMissingPort is an error for an incomplete configuration missing the port
var ErrConfigMissingPort = errors.New("Port is empty")

// ErrConfigMissingName is an error for an incomplete configuration missing the name
var ErrConfigMissingName = errors.New("Name is empty")

// ErrConfigMissingUser is an error for an incomplete configuration missing the user
var ErrConfigMissingUser = errors.New("User is empty")

func validateConfig(c Config) error {
	if c.Host == "" {
		return ErrConfigMissingHost
	}
	if c.Port == "" {
		return ErrConfigMissingPort
	}
	if c.Name == "" {
		return ErrConfigMissingName
	}
	if c.User == "" {
		return ErrConfigMissingUser
	}

	return nil
}

// checkSSLMode checks if SSL is required for the database connection
func checkSSLMode() bool {
	// TODO: deprecate DB_NOSSL in favor of DBSkipSSL
	if os.Getenv("DB_NOSSL") != "" {
		return true
	}

	if os.Getenv("DBSkipSSL") == "true" {
		return true
	}

	return os.Getenv("GO_ENV") != "PRODUCTION"
}

func getPGConnectionString(c Config) (string, error) {
	if err := validateConfig(c); err != nil {
		return "", errors.Wrap(err, "invalid database config")
	}

	var sslmode string
	if checkSSLMode() {
		sslmode = "disable"
	} else {
		sslmode = "require"
	}

	return fmt.Sprintf(
		"sslmode=%s host=%s port=%s dbname=%s user=%s password=%s",
		sslmode,
		c.Host,
		c.Port,
		c.Name,
		c.User,
		c.Password,
	), nil
}

// Open opens the connection with the database
func Open(c Config) *gorm.DB {
	connStr, err := getPGConnectionString(c)
	if err != nil {
		panic(errors.Wrap(err, "getting connection string"))
	}

	conn, err := gorm.Open("postgres", connStr)
	if err != nil {
		panic(errors.Wrap(err, "opening database connection"))
	}

	return conn
}
