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

package presenters

import (
	"time"

	"github.com/nadproject/nad/pkg/server/database"
)

// Book is a result of PresentBooks
type Book struct {
	UUID      string    `json:"uuid"`
	USN       int       `json:"usn"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Label     string    `json:"label"`
}

// PresentBook presents a book
func PresentBook(book database.Book) Book {
	return Book{
		UUID:      book.UUID,
		USN:       book.USN,
		CreatedAt: FormatTS(book.CreatedAt),
		UpdatedAt: FormatTS(book.UpdatedAt),
		Label:     book.Label,
	}
}

// PresentBooks presents books
func PresentBooks(books []database.Book) []Book {
	ret := []Book{}

	for _, book := range books {
		p := PresentBook(book)
		ret = append(ret, p)
	}

	return ret
}
