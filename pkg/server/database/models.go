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
	UUID      string `json:"uuid" gorm:"index;type:uuid;default:uuid_generate_v4()"`
	UserID    int    `json:"user_id" gorm:"index"`
	Label     string `json:"label" gorm:"index"`
	Notes     []Note `json:"notes" gorm:"foreignkey:book_uuid"`
	AddedOn   int64  `json:"added_on"`
	EditedOn  int64  `json:"edited_on"`
	USN       int    `json:"-" gorm:"index"`
	Deleted   bool   `json:"-" gorm:"default:false"`
	Encrypted bool   `json:"-" gorm:"default:false"`
}

// Note is a model for a note
type Note struct {
	Model
	UUID       string     `json:"uuid" gorm:"index;type:uuid;default:uuid_generate_v4()"`
	Book       Book       `json:"book" gorm:"foreignkey:BookUUID"`
	User       User       `json:"user"`
	UserID     int        `json:"user_id" gorm:"index"`
	BookUUID   string     `json:"book_uuid" gorm:"index;type:uuid"`
	Body       string     `json:"content"`
	AddedOn    int64      `json:"added_on"`
	EditedOn   int64      `json:"edited_on"`
	TSV        string     `json:"-" gorm:"type:tsvector"`
	Public     bool       `json:"public" gorm:"default:false"`
	USN        int        `json:"-" gorm:"index"`
	Deleted    bool       `json:"-" gorm:"default:false"`
	Encrypted  bool       `json:"-" gorm:"default:false"`
	NoteReview NoteReview `json:"-"`
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
	APIKey           string     `json:"-" gorm:"index"`                 // Deprecated
	Name             string     `json:"name"`                           // Deprecated
	Encrypted        bool       `json:"encrypted" gorm:"default:false"` // Deprecated
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

// EmailPreference is a preference per user for receiving email communication
type EmailPreference struct {
	Model
	UserID           int  `gorm:"index" json:"-"`
	InactiveReminder bool `json:"inactive_reminder" gorm:"default:true"`
	ProductUpdate    bool `json:"product_update" gorm:"default:true"`
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
	Model
	UUID     string          `json:"uuid" gorm:"type:uuid;index;default:uuid_generate_v4()"`
	RuleID   int             `gorm:"index"`
	Rule     RepetitionRule  `json:"rule"`
	UserID   int             `gorm:"index"`
	Version  int             `gorm:"version"`
	Notes    []Note          `gorm:"many2many:digest_notes;"`
	Receipts []DigestReceipt `gorm:"polymorphic:Target;"`
}

// DigestNote is an intermediary to represent many-to-many relationship
// between digests and notes
type DigestNote struct {
	Model
	NoteID   int `gorm:"index"`
	DigestID int `gorm:"index"`
}

// RepetitionRule is the rules for sending digest emails
type RepetitionRule struct {
	Model
	UUID    string `json:"uuid" gorm:"type:uuid;index;default:uuid_generate_v4()"`
	UserID  int    `json:"user_id" gorm:"index"`
	Title   string `json:"title"`
	Enabled bool   `json:"enabled"`
	Hour    int    `json:"hour" gorm:"index"`
	Minute  int    `json:"minute" gorm:"index"`
	// in milliseconds
	Frequency int64 `json:"frequency"`
	// in milliseconds
	LastActive int64 `json:"last_active"`
	// in milliseconds
	NextActive int64  `json:"next_active"`
	BookDomain string `json:"book_domain"`
	Books      []Book `gorm:"many2many:repetition_rule_books;"`
	NoteCount  int    `json:"note_count"`
}

// DigestReceipt is a read receipt for digests
type DigestReceipt struct {
	Model
	UserID   int `json:"user_id" gorm:"index"`
	DigestID int `json:"digest_id" gorm:"index"`
}

// NoteReview is a record for reviewing a note in a digest
type NoteReview struct {
	Model
	UUID     string `json:"uuid" gorm:"index;type:uuid;default:uuid_generate_v4()"`
	UserID   int    `json:"user_id" gorm:"index"`
	DigestID int    `json:"digest_id" gorm:"index"`
	NoteID   int    `json:"note_id" gorm:"index"`
}
