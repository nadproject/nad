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
	"github.com/nadproject/nad/pkg/server/database"
)

// ViewNote checks if the given user can view the given note
func ViewNote(user *database.User, note database.Note) bool {
	if note.Public {
		return true
	}
	if user == nil {
		return false
	}
	if note.UserID == 0 {
		return false
	}

	return note.UserID == user.ID
}
