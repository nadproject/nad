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

package permissions

import (
	"testing"

	"github.com/nadproject/nad/pkg/assert"
	"github.com/nadproject/nad/pkg/server/models"
)

func TestViewNote(t *testing.T) {
	user := models.User{Model: models.Model{ID: 1}}
	anotherUser := models.User{Model: models.Model{ID: 2}}

	book := models.Book{
		UserID: user.ID,
		Name:   "js",
	}
	privateNote := models.Note{
		UserID:   user.ID,
		BookUUID: book.UUID,
		Deleted:  false,
		Public:   false,
	}
	publicNote := models.Note{
		UserID:   user.ID,
		BookUUID: book.UUID,
		Deleted:  false,
		Public:   true,
	}
	deletedNote := models.Note{
		UserID:   user.ID,
		BookUUID: book.UUID,
		Deleted:  true,
	}

	t.Run("owner viewing private note", func(t *testing.T) {
		result := ViewNote(user.ID, privateNote)
		assert.Equal(t, result, true, "result mismatch")
	})

	t.Run("owner viewing public note", func(t *testing.T) {
		result := ViewNote(user.ID, publicNote)
		assert.Equal(t, result, true, "result mismatch")
	})

	t.Run("owner viewing deleted note", func(t *testing.T) {
		result := ViewNote(user.ID, deletedNote)
		assert.Equal(t, result, false, "result mismatch")
	})

	t.Run("non-owner viewing private note", func(t *testing.T) {
		result := ViewNote(anotherUser.ID, privateNote)
		assert.Equal(t, result, false, "result mismatch")
	})

	t.Run("non-owner viewing public note", func(t *testing.T) {
		result := ViewNote(anotherUser.ID, publicNote)
		assert.Equal(t, result, true, "result mismatch")
	})

	t.Run("non-owner viewing deleted note", func(t *testing.T) {
		result := ViewNote(anotherUser.ID, deletedNote)
		assert.Equal(t, result, false, "result mismatch")
	})

	t.Run("guest viewing private note", func(t *testing.T) {
		result := ViewNote(0, privateNote)
		assert.Equal(t, result, false, "result mismatch")
	})

	t.Run("guest viewing public note", func(t *testing.T) {
		result := ViewNote(0, publicNote)
		assert.Equal(t, result, true, "result mismatch")
	})

	t.Run("guest viewing deleted note", func(t *testing.T) {
		result := ViewNote(0, deletedNote)
		assert.Equal(t, result, false, "result mismatch")
	})
}

func TestUpdateNote(t *testing.T) {
	user := models.User{Model: models.Model{ID: 1}}
	anotherUser := models.User{Model: models.Model{ID: 2}}

	book := models.Book{
		UserID: user.ID,
		Name:   "js",
	}
	note := models.Note{
		UserID:   user.ID,
		BookUUID: book.UUID,
		Deleted:  false,
	}

	t.Run("owner updating note", func(t *testing.T) {
		result := UpdateNote(user.ID, note)
		assert.Equal(t, result, true, "result mismatch")
	})

	t.Run("non-owner updating note", func(t *testing.T) {
		result := UpdateNote(anotherUser.ID, note)
		assert.Equal(t, result, false, "result mismatch")
	})

	t.Run("guest updating note", func(t *testing.T) {
		result := UpdateNote(0, note)
		assert.Equal(t, result, false, "result mismatch")
	})
}

func TestDeleteNote(t *testing.T) {
	user := models.User{Model: models.Model{ID: 1}}
	anotherUser := models.User{Model: models.Model{ID: 2}}

	book := models.Book{
		UserID: user.ID,
		Name:   "js",
	}
	note := models.Note{
		UserID:   user.ID,
		BookUUID: book.UUID,
		Deleted:  false,
	}

	t.Run("owner deleting note", func(t *testing.T) {
		result := DeleteNote(user.ID, note)
		assert.Equal(t, result, true, "result mismatch")
	})

	t.Run("non-owner deleting note", func(t *testing.T) {
		result := DeleteNote(anotherUser.ID, note)
		assert.Equal(t, result, false, "result mismatch")
	})

	t.Run("guest deleting note", func(t *testing.T) {
		result := DeleteNote(0, note)
		assert.Equal(t, result, false, "result mismatch")
	})
}

func TestViewBook(t *testing.T) {
	user := models.User{Model: models.Model{ID: 1}}
	anotherUser := models.User{Model: models.Model{ID: 2}}

	book := models.Book{
		UserID: user.ID,
		Name:   "js",
	}

	t.Run("owner viewing book", func(t *testing.T) {
		result := ViewBook(user.ID, book)
		assert.Equal(t, result, true, "result mismatch")
	})

	t.Run("non-owner viewing book", func(t *testing.T) {
		result := ViewBook(anotherUser.ID, book)
		assert.Equal(t, result, false, "result mismatch")
	})

	t.Run("guest viewing book", func(t *testing.T) {
		result := ViewBook(0, book)
		assert.Equal(t, result, false, "result mismatch")
	})
}

func TestUpdateBook(t *testing.T) {
	user := models.User{Model: models.Model{ID: 1}}
	anotherUser := models.User{Model: models.Model{ID: 2}}

	book := models.Book{
		UserID: user.ID,
		Name:   "js",
	}

	t.Run("owner updating book", func(t *testing.T) {
		result := UpdateBook(user.ID, book)
		assert.Equal(t, result, true, "result mismatch")
	})

	t.Run("non-owner updating book", func(t *testing.T) {
		result := UpdateBook(anotherUser.ID, book)
		assert.Equal(t, result, false, "result mismatch")
	})

	t.Run("guest updating book", func(t *testing.T) {
		result := UpdateBook(0, book)
		assert.Equal(t, result, false, "result mismatch")
	})
}

func TestDeleteBook(t *testing.T) {
	user := models.User{Model: models.Model{ID: 1}}
	anotherUser := models.User{Model: models.Model{ID: 2}}

	book := models.Book{
		UserID: user.ID,
		Name:   "js",
	}

	t.Run("owner deleting book", func(t *testing.T) {
		result := DeleteBook(user.ID, book)
		assert.Equal(t, result, true, "result mismatch")
	})

	t.Run("non-owner deleting book", func(t *testing.T) {
		result := DeleteBook(anotherUser.ID, book)
		assert.Equal(t, result, false, "result mismatch")
	})

	t.Run("guest deleting book", func(t *testing.T) {
		result := DeleteBook(0, book)
		assert.Equal(t, result, false, "result mismatch")
	})
}
