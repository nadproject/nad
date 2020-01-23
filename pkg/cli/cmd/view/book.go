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
	"database/sql"
	"fmt"
	"strings"

	"github.com/nadproject/nad/pkg/cli/context"
	"github.com/nadproject/nad/pkg/cli/log"
	"github.com/pkg/errors"
)

// bookInfo is an information about the book to be printed on screen
type bookInfo struct {
	BookLabel string
	NoteCount int
}

// noteInfo is an information about the note to be printed on screen
type noteInfo struct {
	RowID int
	Body  string
}

// getNewlineIdx returns the index of newline character in a string
func getNewlineIdx(str string) int {
	var ret int

	ret = strings.Index(str, "\n")

	if ret == -1 {
		ret = strings.Index(str, "\r\n")
	}

	return ret
}

// formatBody returns an excerpt of the given raw note content and a boolean
// indicating if the returned string has been excertped
func formatBody(noteBody string) (string, bool) {
	trimmed := strings.TrimRight(noteBody, "\r\n")
	newlineIdx := getNewlineIdx(trimmed)

	if newlineIdx > -1 {
		ret := strings.Trim(trimmed[0:newlineIdx], " ")

		return ret, true
	}

	return strings.Trim(trimmed, " "), false
}

func printBookLine(info bookInfo, nameOnly bool) {
	if nameOnly {
		fmt.Println(info.BookLabel)
	} else {
		log.Printf("%s %s\n", info.BookLabel, log.ColorYellow.Sprintf("(%d)", info.NoteCount))
	}
}

func printBooks(ctx context.NadCtx, nameOnly bool) error {
	db := ctx.DB

	rows, err := db.Query(`SELECT books.name, count(notes.uuid) note_count
	FROM books
	LEFT JOIN notes ON notes.book_uuid = books.uuid AND notes.deleted = false
	WHERE books.deleted = false
	GROUP BY books.uuid
	ORDER BY books.name ASC;`)
	if err != nil {
		return errors.Wrap(err, "querying books")
	}
	defer rows.Close()

	infos := []bookInfo{}
	for rows.Next() {
		var info bookInfo
		err = rows.Scan(&info.BookLabel, &info.NoteCount)
		if err != nil {
			return errors.Wrap(err, "scanning a row")
		}

		infos = append(infos, info)
	}

	for _, info := range infos {
		printBookLine(info, nameOnly)
	}

	return nil
}

func printBookNotes(ctx context.NadCtx, bookName string) error {
	db := ctx.DB

	var bookUUID string
	err := db.QueryRow("SELECT uuid FROM books WHERE name = ?", bookName).Scan(&bookUUID)
	if err == sql.ErrNoRows {
		return errors.New("book not found")
	} else if err != nil {
		return errors.Wrap(err, "querying the book")
	}

	rows, err := db.Query(`SELECT rowid, body FROM notes WHERE book_uuid = ? AND deleted = ? ORDER BY added_on ASC;`, bookUUID, false)
	if err != nil {
		return errors.Wrap(err, "querying notes")
	}
	defer rows.Close()

	infos := []noteInfo{}
	for rows.Next() {
		var info noteInfo
		err = rows.Scan(&info.RowID, &info.Body)
		if err != nil {
			return errors.Wrap(err, "scanning a row")
		}

		infos = append(infos, info)
	}

	log.Infof("on book %s\n", bookName)

	for _, info := range infos {
		body, isExcerpt := formatBody(info.Body)

		rowid := log.ColorYellow.Sprintf("(%d)", info.RowID)
		if isExcerpt {
			body = fmt.Sprintf("%s %s", body, log.ColorYellow.Sprintf("[---More---]"))
		}

		log.Plainf("%s %s\n", rowid, body)
	}

	return nil
}
