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

package config

import (
	"fmt"
	"net/url"
	"os"

	"github.com/pkg/errors"
)

const (
	// AppEnvProduction represents an app environment for production.
	AppEnvProduction string = "PRODUCTION"
)

var (
	// ErrDBMissingHost is an error for an incomplete configuration missing the host
	ErrDBMissingHost = errors.New("Host is empty")
	// ErrDBMissingPort is an error for an incomplete configuration missing the port
	ErrDBMissingPort = errors.New("Port is empty")
	// ErrDBMissingName is an error for an incomplete configuration missing the name
	ErrDBMissingName = errors.New("Name is empty")
	// ErrDBMissingUser is an error for an incomplete configuration missing the user
	ErrDBMissingUser = errors.New("User is empty")
	// ErrWebURLInvalid is an error for an incomplete configuration missing the user
	ErrWebURLInvalid = errors.New("invalid WebURL")
)

// PostgresConfig holds the postgres connection configuration.
type PostgresConfig struct {
	SSLMode  string
	Host     string
	Port     string
	Name     string
	User     string
	Password string
}

// Config holds the application configuration
type Config struct {
	AppEnv              string
	Port                string
	WebURL              string
	OnPremise           bool
	DisableRegistration bool
	DB                  PostgresConfig
}

func readBoolEnv(name string) bool {
	if os.Getenv(name) == "true" {
		return true
	}

	return false
}

func loadDBConfig() PostgresConfig {
	var sslmode string
	if readBoolEnv("DB_SKIP_SSL") {
		sslmode = "disable"
	} else {
		sslmode = "require"
	}

	return PostgresConfig{
		SSLMode:  sslmode,
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		Name:     os.Getenv("DB_NAME"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
	}
}

// Load constructs and returns a new config based on the environment variables.
func Load() Config {
	c := Config{
		AppEnv:              os.Getenv("APP_ENV"),
		WebURL:              os.Getenv("WEB_URL"),
		OnPremise:           readBoolEnv("ON_PREMISE"),
		DisableRegistration: readBoolEnv("DISABLE_REGISTRATION"),
		DB:                  loadDBConfig(),
	}

	if err := validate(c); err != nil {
		panic(err)
	}

	return c
}

func validate(c Config) error {
	if _, err := url.ParseRequestURI(c.WebURL); err != nil {
		return ErrWebURLInvalid
	}

	if c.DB.Host == "" {
		return ErrDBMissingHost
	}
	if c.DB.Port == "" {
		return ErrDBMissingPort
	}
	if c.DB.Name == "" {
		return ErrDBMissingName
	}
	if c.DB.User == "" {
		return ErrDBMissingUser
	}

	return nil
}

// GetConnectionStr returns a postgres connection string.
func (c PostgresConfig) GetConnectionStr() string {
	return fmt.Sprintf(
		"sslmode=%s host=%s port=%s dbname=%s user=%s password=%s",
		c.SSLMode, c.Host, c.Port, c.Name, c.User, c.Password)
}
