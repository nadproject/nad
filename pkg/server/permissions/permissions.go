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
	"github.com/nadproject/nad/pkg/server/models"
)

func isOwner(userID uint, note models.Note) bool {
	return note.UserID == userID
}

// ViewNote checks if the given user can view the given note
func ViewNote(userID uint, note models.Note) bool {
	if note.Public {
		return true
	}

	return isOwner(userID, note)
}

// UpdateNote checks if the given user can update the given note
func UpdateNote(userID uint, note models.Note) bool {
	return isOwner(userID, note)
}

// DeleteNote checks if the given user can delete the given note
func DeleteNote(userID uint, note models.Note) bool {
	return isOwner(userID, note)
}
