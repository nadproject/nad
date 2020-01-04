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

package app

import (
	"github.com/nadproject/nad/pkg/clock"
	"github.com/nadproject/nad/pkg/server/mailer"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/stripe/stripe-go"
)

var (
	// ErrEmptyDB is an error for missing database connection in the app configuration
	ErrEmptyDB = errors.New("No database connection was provided")
	// ErrEmptyClock is an error for missing clock in the app configuration
	ErrEmptyClock = errors.New("No clock was provided")
	// ErrEmptyWebURL is an error for missing WebURL content in the app configuration
	ErrEmptyWebURL = errors.New("No WebURL was provided")
	// ErrEmptyEmailTemplates is an error for missing EmailTemplates content in the app configuration
	ErrEmptyEmailTemplates = errors.New("No EmailTemplate store was provided")
	// ErrEmptyEmailBackend is an error for missing EmailBackend content in the app configuration
	ErrEmptyEmailBackend = errors.New("No EmailBackend was provided")
)

// Config is an application configuration
type Config struct {
	WebURL              string
	OnPremise           bool
	DisableRegistration bool
}

// App is an application context
type App struct {
	DB               *gorm.DB
	Clock            clock.Clock
	StripeAPIBackend stripe.Backend
	EmailTemplates   mailer.Templates
	EmailBackend     mailer.Backend
	Config           Config
}

// Validate validates the app configuration
func (a *App) Validate() error {
	if a.Config.WebURL == "" {
		return ErrEmptyWebURL
	}
	if a.Clock == nil {
		return ErrEmptyClock
	}
	if a.EmailTemplates == nil {
		return ErrEmptyEmailTemplates
	}
	if a.EmailBackend == nil {
		return ErrEmptyEmailBackend
	}
	if a.DB == nil {
		return ErrEmptyDB
	}

	return nil
}
