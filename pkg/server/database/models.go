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

package database

import (
	"time"
)

// Model is the base model definition
type Model struct {
	ID        int       `gorm:"primary_key" json:"-"`
	CreatedAt time.Time `json:"created_at" gorm:"default:now()"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Book is a model for a book
type Book struct {
	Model
	UUID     string `json:"uuid" gorm:"index;type:uuid;default:uuid_generate_v4()"`
	UserID   int    `json:"user_id" gorm:"index"`
	Label    string `json:"label" gorm:"index"`
	Notes    []Note `json:"notes" gorm:"foreignkey:book_uuid"`
	AddedOn  int64  `json:"added_on"`
	EditedOn int64  `json:"edited_on"`
	USN      int    `json:"-" gorm:"index"`
	Deleted  bool   `json:"-" gorm:"default:false"`
}

// Note is a model for a note
type Note struct {
	Model
	UUID     string `json:"uuid" gorm:"index;type:uuid;default:uuid_generate_v4()"`
	Book     Book   `json:"book" gorm:"foreignkey:BookUUID"`
	User     User   `json:"user"`
	UserID   int    `json:"user_id" gorm:"index"`
	BookUUID string `json:"book_uuid" gorm:"index;type:uuid"`
	Body     string `json:"content"`
	AddedOn  int64  `json:"added_on"`
	EditedOn int64  `json:"edited_on"`
	TSV      string `json:"-" gorm:"type:tsvector"`
	Public   bool   `json:"public" gorm:"default:false"`
	USN      int    `json:"-" gorm:"index"`
	Deleted  bool   `json:"-" gorm:"default:false"`
}

// User is a model for a user
type User struct {
	Model
	UUID             string `json:"uuid" gorm:"type:uuid;index;default:uuid_generate_v4()"`
	StripeCustomerID string `json:"-"`
	BillingCountry   string `json:"-"`
	Account          Account
	LastLoginAt      *time.Time `json:"-"`
	MaxUSN           int        `json:"-" gorm:"default:0"`
	Cloud            bool       `json:"-" gorm:"default:false"`
}

// Account is a model for an account
type Account struct {
	Model
	UserID             int    `gorm:"index"`
	AccountID          string // Deprecated
	Nickname           string // Deprecated
	Provider           string // Deprecated
	Email              NullString
	EmailVerified      bool       `gorm:"default:false"`
	Password           NullString // Deprecated
	ClientKDFIteration int
	ServerKDFIteration int
	AuthKeyHash        string
	Salt               string
	CipherKeyEnc       string
}

// Token is a model for a token
type Token struct {
	Model
	UserID int    `gorm:"index"`
	Value  string `gorm:"index"`
	Type   string
	UsedAt *time.Time
}

// Notification is the learning notification sent to the user
type Notification struct {
	Model
	Type   string
	UserID int `gorm:"index"`
}

// EmailPreference is information about how often user wants to receive digest email
type EmailPreference struct {
	Model
	UserID       int  `gorm:"index" json:"-"`
	DigestWeekly bool `json:"digest_weekly"`
}

// Session represents a user session
type Session struct {
	Model
	UserID     int    `gorm:"index"`
	Key        string `gorm:"index"`
	LastUsedAt time.Time
	ExpiresAt  time.Time
}

// Digest is a digest of notes
type Digest struct {
	UUID      string    `json:"uuid" gorm:"primary_key:true;type:uuid;index;default:uuid_generate_v4()"`
	UserID    int       `gorm:"index"`
	Notes     []Note    `gorm:"many2many:digest_notes;association_foreignKey:uuid;association_jointable_foreignkey:note_uuid;jointable_foreignkey:digest_uuid;"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DigestRule is the rules for sending digest emails
type DigestRule struct {
	Model
	UUID       string `json:"uuid" gorm:"type:uuid;index;default:uuid_generate_v4()"`
	UserID     int    `json:"user_id" gorm:"index"`
	Title      string `json:"title"`
	Enabled    bool   `json:"enabled"`
	Hour       int    `json:"hour" gorm:"index"`
	Minute     int    `json:"minute" gorm:"index"`
	Frequency  int    `json:"frequency"`
	LastActive int    `json:"-"`
	Books      []Book `gorm:"many2many:repetition_rule_books;"`
	NoteCount  int    `json:"note_count"`
}
