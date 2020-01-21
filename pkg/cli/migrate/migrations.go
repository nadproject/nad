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

package migrate

import (
	"github.com/nadproject/nad/pkg/cli/context"
	"github.com/nadproject/nad/pkg/cli/database"
	"github.com/pkg/errors"
)

type migration struct {
	name string
	run  func(ctx context.NadCtx, tx *database.DB) error
}

var lm1 = migration{
	name: "initialize schema",
	run: func(ctx context.NadCtx, tx *database.DB) error {
		_, err := tx.Exec(`CREATE TABLE IF NOT EXISTS notes
		(
			uuid text NOT NULL,
			book_uuid text NOT NULL,
			body text NOT NULL,
			added_on integer NOT NULL,
			edited_on integer DEFAULT 0,
			public bool DEFAULT false,
			dirty bool DEFAULT false,
			usn int DEFAULT 0 NOT NULL,
			deleted bool DEFAULT false
		)`)
		if err != nil {
			return errors.Wrap(err, "creating notes table")
		}

		_, err = tx.Exec(`CREATE TABLE IF NOT EXISTS books
		(
			uuid text PRIMARY KEY,
			name text NOT NULL,
			dirty bool DEFAULT false,
			usn int DEFAULT 0 NOT NULL,
			deleted bool DEFAULT false
		)`)
		if err != nil {
			return errors.Wrap(err, "creating books table")
		}

		_, err = tx.Exec(`CREATE TABLE IF NOT EXISTS system
		(
			key string NOT NULL,
			value text NOT NULL
		)`)
		if err != nil {
			return errors.Wrap(err, "creating system table")
		}

		_, err = tx.Exec(`CREATE TABLE IF NOT EXISTS actions
		(
			uuid text PRIMARY KEY,
			schema integer NOT NULL,
			type text NOT NULL,
			data text NOT NULL,
			timestamp integer NOT NULL
		)`)
		if err != nil {
			return errors.Wrap(err, "creating actions table")
		}

		_, err = tx.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_books_name ON books(name);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_notes_uuid ON notes(uuid);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_books_uuid ON books(uuid);
		CREATE INDEX IF NOT EXISTS idx_notes_book_uuid ON notes(book_uuid);`)
		if err != nil {
			return errors.Wrap(err, "creating indices")
		}

		return nil
	},
}
