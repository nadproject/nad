/* Copyright (C) 2019 Monomax Software Pty Ltd
 *
 * This file is part of NAD.
 *
 * NAD is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * NAD is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with NAD.  If not, see <https://www.gnu.org/licenses/>.
 */

package view

import (
	"strconv"

	"github.com/nadproject/nad/pkg/cli/context"
	"github.com/nadproject/nad/pkg/cli/database"
	"github.com/nadproject/nad/pkg/cli/output"
	"github.com/pkg/errors"
)

func printNote(ctx context.NadCtx, rowID string) error {
	noteRowID, err := strconv.Atoi(rowID)
	if err != nil {
		return errors.Wrap(err, "invalid rowid")
	}

	db := ctx.DB
	info, err := database.GetNoteInfo(db, noteRowID)
	if err != nil {
		return err
	}

	output.NoteInfo(info)

	return nil
}
