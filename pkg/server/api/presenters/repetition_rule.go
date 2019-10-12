/* Copyright (C) 2019 Monomax Software Pty Ltd
 *
 * This file is part of Dnote.
 *
 * Dnote is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * Dnote is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with Dnote.  If not, see <https://www.gnu.org/licenses/>.
 */

package presenters

import (
	"time"

	"github.com/nadproject/nad/pkg/server/database"
)

// DigestRule is a presented digest rule
type DigestRule struct {
	UUID      string    `json:"uuid"`
	Title     string    `json:"title"`
	Enabled   bool      `json:"enabled"`
	Hour      int       `json:"hour" gorm:"index"`
	Minute    int       `json:"minute" gorm:"index"`
	Frequency int       `json:"frequency"`
	Books     []Book    `json:"books"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PresentDigestRule presents a digest rule
func PresentDigestRule(d database.DigestRule) DigestRule {
	ret := DigestRule{
		UUID:      d.UUID,
		Title:     d.Title,
		Enabled:   d.Enabled,
		Hour:      d.Hour,
		Minute:    d.Minute,
		Frequency: d.Frequency,
		Books:     PresentBooks(d.Books),
		CreatedAt: FormatTS(d.CreatedAt),
		UpdatedAt: FormatTS(d.UpdatedAt),
	}

	return ret
}

// PresentDigestRules presents a slice of digest rules
func PresentDigestRules(ds []database.DigestRule) []DigestRule {
	ret := []DigestRule{}

	for _, d := range ds {
		p := PresentDigestRule(d)
		ret = append(ret, p)
	}

	return ret
}
