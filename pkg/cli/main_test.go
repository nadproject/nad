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

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"testing"

	"github.com/nadproject/nad/pkg/assert"
	"github.com/nadproject/nad/pkg/cli/consts"
	"github.com/nadproject/nad/pkg/cli/database"
	"github.com/nadproject/nad/pkg/cli/testutils"
	"github.com/nadproject/nad/pkg/cli/utils"
	"github.com/pkg/errors"
)

var binaryName = "test-nad"

var opts = testutils.RunNADCmdOptions{
	HomeDir: "./tmp",
	NADDir:  "./tmp/.nad",
}

func TestMain(m *testing.M) {
	if err := exec.Command("go", "build", "--tags", "fts5", "-o", binaryName).Run(); err != nil {
		log.Print(errors.Wrap(err, "building a binary").Error())
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func TestInit(t *testing.T) {
	// Execute
	testutils.RunNADCmd(t, opts, binaryName)
	defer testutils.RemoveDir(t, opts.HomeDir)

	db := database.OpenTestDB(t, opts.NADDir)

	// Test
	ok, err := utils.FileExists(opts.NADDir)
	if err != nil {
		t.Fatal(errors.Wrap(err, "checking if nad dir exists"))
	}
	if !ok {
		t.Errorf("nad directory was not initialized")
	}

	ok, err = utils.FileExists(fmt.Sprintf("%s/%s", opts.NADDir, consts.ConfigFilename))
	if err != nil {
		t.Fatal(errors.Wrap(err, "checking if nad config exists"))
	}
	if !ok {
		t.Errorf("config file was not initialized")
	}

	var notesTableCount, booksTableCount, systemTableCount int
	database.MustScan(t, "counting notes",
		db.QueryRow("SELECT count(*) FROM sqlite_master WHERE type = ? AND name = ?", "table", "notes"), &notesTableCount)
	database.MustScan(t, "counting books",
		db.QueryRow("SELECT count(*) FROM sqlite_master WHERE type = ? AND name = ?", "table", "books"), &booksTableCount)
	database.MustScan(t, "counting system",
		db.QueryRow("SELECT count(*) FROM sqlite_master WHERE type = ? AND name = ?", "table", "system"), &systemTableCount)

	assert.Equal(t, notesTableCount, 1, "notes table count mismatch")
	assert.Equal(t, booksTableCount, 1, "books table count mismatch")
	assert.Equal(t, systemTableCount, 1, "system table count mismatch")

	// test that all default system configurations are generated
	var lastUpgrade, lastMaxUSN, lastSyncAt string
	database.MustScan(t, "scanning last upgrade",
		db.QueryRow("SELECT value FROM system WHERE key = ?", consts.SystemLastUpgrade), &lastUpgrade)
	database.MustScan(t, "scanning last max usn",
		db.QueryRow("SELECT value FROM system WHERE key = ?", consts.SystemLastMaxUSN), &lastMaxUSN)
	database.MustScan(t, "scanning last sync at",
		db.QueryRow("SELECT value FROM system WHERE key = ?", consts.SystemLastSyncAt), &lastSyncAt)

	assert.NotEqual(t, lastUpgrade, "", "last upgrade should not be empty")
	assert.NotEqual(t, lastMaxUSN, "", "last max usn should not be empty")
	assert.NotEqual(t, lastSyncAt, "", "last sync at should not be empty")
}

func TestAddNote(t *testing.T) {
	t.Run("new book", func(t *testing.T) {
		// Set up and execute
		testutils.RunNADCmd(t, opts, binaryName, "add", "js", "-c", "foo")
		defer testutils.RemoveDir(t, opts.HomeDir)

		db := database.OpenTestDB(t, opts.NADDir)

		// Test
		var noteCount, bookCount int
		database.MustScan(t, "counting books", db.QueryRow("SELECT count(*) FROM books"), &bookCount)
		database.MustScan(t, "counting notes", db.QueryRow("SELECT count(*) FROM notes"), &noteCount)

		assert.Equalf(t, bookCount, 1, "book count mismatch")
		assert.Equalf(t, noteCount, 1, "note count mismatch")

		var book database.Book
		database.MustScan(t, "getting book", db.QueryRow("SELECT uuid, dirty FROM books where name = ?", "js"), &book.UUID, &book.Dirty)
		var note database.Note
		database.MustScan(t, "getting note",
			db.QueryRow("SELECT uuid, body, added_on, dirty FROM notes where book_uuid = ?", book.UUID), &note.UUID, &note.Body, &note.AddedOn, &note.Dirty)

		assert.Equal(t, book.Dirty, true, "Book dirty mismatch")

		assert.NotEqual(t, note.UUID, "", "Note should have UUID")
		assert.Equal(t, note.Body, "foo", "Note body mismatch")
		assert.Equal(t, note.Dirty, true, "Note dirty mismatch")
		assert.NotEqual(t, note.AddedOn, int64(0), "Note added_on mismatch")
	})

	t.Run("existing book", func(t *testing.T) {
		// Setup
		db := database.InitTestDB(t, fmt.Sprintf("%s/%s", opts.NADDir, consts.NADDBFileName), nil)
		testutils.Setup3(t, db)

		// Execute
		testutils.RunNADCmd(t, opts, binaryName, "add", "js", "-c", "foo")
		defer testutils.RemoveDir(t, opts.HomeDir)

		// Test

		var noteCount, bookCount int
		database.MustScan(t, "counting books", db.QueryRow("SELECT count(*) FROM books"), &bookCount)
		database.MustScan(t, "counting notes", db.QueryRow("SELECT count(*) FROM notes"), &noteCount)

		assert.Equalf(t, bookCount, 1, "book count mismatch")
		assert.Equalf(t, noteCount, 2, "note count mismatch")

		var n1, n2 database.Note
		database.MustScan(t, "getting n1",
			db.QueryRow("SELECT uuid, body, added_on, dirty FROM notes WHERE book_uuid = ? AND uuid = ?", "js-book-uuid", "43827b9a-c2b0-4c06-a290-97991c896653"), &n1.UUID, &n1.Body, &n1.AddedOn, &n1.Dirty)
		database.MustScan(t, "getting n2",
			db.QueryRow("SELECT uuid, body, added_on, dirty FROM notes WHERE book_uuid = ? AND body = ?", "js-book-uuid", "foo"), &n2.UUID, &n2.Body, &n2.AddedOn, &n2.Dirty)

		var book database.Book
		database.MustScan(t, "getting book", db.QueryRow("SELECT dirty FROM books where name = ?", "js"), &book.Dirty)

		assert.Equal(t, book.Dirty, false, "Book dirty mismatch")

		assert.NotEqual(t, n1.UUID, "", "n1 should have UUID")
		assert.Equal(t, n1.Body, "Booleans have toString()", "n1 body mismatch")
		assert.Equal(t, n1.AddedOn, int64(1515199943), "n1 added_on mismatch")
		assert.Equal(t, n1.Dirty, false, "n1 dirty mismatch")

		assert.NotEqual(t, n2.UUID, "", "n2 should have UUID")
		assert.Equal(t, n2.Body, "foo", "n2 body mismatch")
		assert.Equal(t, n2.Dirty, true, "n2 dirty mismatch")
	})
}

func TestEditNote(t *testing.T) {
	t.Run("content flag", func(t *testing.T) {
		// Setup
		db := database.InitTestDB(t, fmt.Sprintf("%s/%s", opts.NADDir, consts.NADDBFileName), nil)
		testutils.Setup4(t, db)

		// Execute
		testutils.RunNADCmd(t, opts, binaryName, "edit", "2", "-c", "foo bar")
		defer testutils.RemoveDir(t, opts.HomeDir)

		// Test
		var noteCount, bookCount int
		database.MustScan(t, "counting books", db.QueryRow("SELECT count(*) FROM books"), &bookCount)
		database.MustScan(t, "counting notes", db.QueryRow("SELECT count(*) FROM notes"), &noteCount)

		assert.Equalf(t, bookCount, 1, "book count mismatch")
		assert.Equalf(t, noteCount, 2, "note count mismatch")

		var n1, n2 database.Note
		database.MustScan(t, "getting n1",
			db.QueryRow("SELECT uuid, body, added_on, dirty FROM notes where book_uuid = ? AND uuid = ?", "js-book-uuid", "43827b9a-c2b0-4c06-a290-97991c896653"), &n1.UUID, &n1.Body, &n1.AddedOn, &n1.Dirty)
		database.MustScan(t, "getting n2",
			db.QueryRow("SELECT uuid, body, added_on, dirty FROM notes where book_uuid = ? AND uuid = ?", "js-book-uuid", "f0d0fbb7-31ff-45ae-9f0f-4e429c0c797f"), &n2.UUID, &n2.Body, &n2.AddedOn, &n2.Dirty)

		assert.Equal(t, n1.UUID, "43827b9a-c2b0-4c06-a290-97991c896653", "n1 should have UUID")
		assert.Equal(t, n1.Body, "Booleans have toString()", "n1 body mismatch")
		assert.Equal(t, n1.Dirty, false, "n1 dirty mismatch")

		assert.Equal(t, n2.UUID, "f0d0fbb7-31ff-45ae-9f0f-4e429c0c797f", "Note should have UUID")
		assert.Equal(t, n2.Body, "foo bar", "Note body mismatch")
		assert.Equal(t, n2.Dirty, true, "n2 dirty mismatch")
		assert.NotEqual(t, n2.EditedOn, 0, "Note edited_on mismatch")
	})

	t.Run("book flag", func(t *testing.T) {
		// Setup
		db := database.InitTestDB(t, fmt.Sprintf("%s/%s", opts.NADDir, consts.NADDBFileName), nil)
		testutils.Setup5(t, db)

		// Execute
		testutils.RunNADCmd(t, opts, binaryName, "edit", "2", "-b", "linux")
		defer testutils.RemoveDir(t, opts.HomeDir)

		// Test
		var noteCount, bookCount int
		database.MustScan(t, "counting books", db.QueryRow("SELECT count(*) FROM books"), &bookCount)
		database.MustScan(t, "counting notes", db.QueryRow("SELECT count(*) FROM notes"), &noteCount)

		assert.Equalf(t, bookCount, 2, "book count mismatch")
		assert.Equalf(t, noteCount, 2, "note count mismatch")

		var n1, n2 database.Note
		database.MustScan(t, "getting n1",
			db.QueryRow("SELECT uuid, book_uuid, body, added_on, dirty FROM notes where uuid = ?", "f0d0fbb7-31ff-45ae-9f0f-4e429c0c797f"), &n1.UUID, &n1.BookUUID, &n1.Body, &n1.AddedOn, &n1.Dirty)
		database.MustScan(t, "getting n2",
			db.QueryRow("SELECT uuid, book_uuid, body, added_on, dirty FROM notes where uuid = ?", "43827b9a-c2b0-4c06-a290-97991c896653"), &n2.UUID, &n2.BookUUID, &n2.Body, &n2.AddedOn, &n2.Dirty)

		assert.Equal(t, n1.BookUUID, "js-book-uuid", "n1 BookUUID mismatch")
		assert.Equal(t, n1.Body, "n1 body", "n1 Body mismatch")
		assert.Equal(t, n1.Dirty, false, "n1 Dirty mismatch")
		assert.Equal(t, n1.EditedOn, int64(0), "n1 EditedOn mismatch")

		assert.Equal(t, n2.BookUUID, "linux-book-uuid", "n2 BookUUID mismatch")
		assert.Equal(t, n2.Body, "n2 body", "n2 Body mismatch")
		assert.Equal(t, n2.Dirty, true, "n2 Dirty mismatch")
		assert.NotEqual(t, n2.EditedOn, 0, "n2 EditedOn mismatch")
	})

	t.Run("book flag and content flag", func(t *testing.T) {
		// Setup
		db := database.InitTestDB(t, fmt.Sprintf("%s/%s", opts.NADDir, consts.NADDBFileName), nil)
		testutils.Setup5(t, db)

		// Execute
		testutils.RunNADCmd(t, opts, binaryName, "edit", "2", "-b", "linux", "-c", "n2 body updated")
		defer testutils.RemoveDir(t, opts.HomeDir)

		// Test
		var noteCount, bookCount int
		database.MustScan(t, "counting books", db.QueryRow("SELECT count(*) FROM books"), &bookCount)
		database.MustScan(t, "counting notes", db.QueryRow("SELECT count(*) FROM notes"), &noteCount)

		assert.Equalf(t, bookCount, 2, "book count mismatch")
		assert.Equalf(t, noteCount, 2, "note count mismatch")

		var n1, n2 database.Note
		database.MustScan(t, "getting n1",
			db.QueryRow("SELECT uuid, book_uuid, body, added_on, dirty FROM notes where uuid = ?", "f0d0fbb7-31ff-45ae-9f0f-4e429c0c797f"), &n1.UUID, &n1.BookUUID, &n1.Body, &n1.AddedOn, &n1.Dirty)
		database.MustScan(t, "getting n2",
			db.QueryRow("SELECT uuid, book_uuid, body, added_on, dirty FROM notes where uuid = ?", "43827b9a-c2b0-4c06-a290-97991c896653"), &n2.UUID, &n2.BookUUID, &n2.Body, &n2.AddedOn, &n2.Dirty)

		assert.Equal(t, n1.BookUUID, "js-book-uuid", "n1 BookUUID mismatch")
		assert.Equal(t, n1.Body, "n1 body", "n1 Body mismatch")
		assert.Equal(t, n1.Dirty, false, "n1 Dirty mismatch")
		assert.Equal(t, n1.EditedOn, int64(0), "n1 EditedOn mismatch")

		assert.Equal(t, n2.BookUUID, "linux-book-uuid", "n2 BookUUID mismatch")
		assert.Equal(t, n2.Body, "n2 body updated", "n2 Body mismatch")
		assert.Equal(t, n2.Dirty, true, "n2 Dirty mismatch")
		assert.NotEqual(t, n2.EditedOn, 0, "n2 EditedOn mismatch")
	})
}

func TestEditBook(t *testing.T) {
	t.Run("name flag", func(t *testing.T) {
		// Setup
		db := database.InitTestDB(t, fmt.Sprintf("%s/%s", opts.NADDir, consts.NADDBFileName), nil)
		testutils.Setup1(t, db)

		// Execute
		testutils.RunNADCmd(t, opts, binaryName, "edit", "js", "-n", "js-edited")
		defer testutils.RemoveDir(t, opts.HomeDir)

		// Test
		var noteCount, bookCount int
		database.MustScan(t, "counting books", db.QueryRow("SELECT count(*) FROM books"), &bookCount)
		database.MustScan(t, "counting notes", db.QueryRow("SELECT count(*) FROM notes"), &noteCount)

		assert.Equalf(t, bookCount, 2, "book count mismatch")
		assert.Equalf(t, noteCount, 1, "note count mismatch")

		var b1, b2 database.Book
		var n1 database.Note
		database.MustScan(t, "getting b1",

			db.QueryRow("SELECT uuid, name, usn, dirty FROM books WHERE uuid = ?", "js-book-uuid"), &b1.UUID, &b1.Name, &b1.USN, &b1.Dirty)
		database.MustScan(t, "getting b2",
			db.QueryRow("SELECT uuid, name, usn, dirty FROM books WHERE uuid = ?", "linux-book-uuid"), &b2.UUID, &b2.Name, &b2.USN, &b2.Dirty)
		database.MustScan(t, "getting n1",
			db.QueryRow("SELECT uuid, body, added_on, deleted, dirty, usn FROM notes WHERE book_uuid = ? AND uuid = ?", "js-book-uuid", "43827b9a-c2b0-4c06-a290-97991c896653"),
			&n1.UUID, &n1.Body, &n1.AddedOn, &n1.Deleted, &n1.Dirty, &n1.USN)

		assert.Equal(t, b1.UUID, "js-book-uuid", "b1 should have UUID")
		assert.Equal(t, b1.Name, "js-edited", "b1 Name mismatch")
		assert.Equal(t, b1.USN, 0, "b1 USN mismatch")
		assert.Equal(t, b1.Dirty, true, "b1 Dirty mismatch")

		assert.Equal(t, b2.UUID, "linux-book-uuid", "b2 should have UUID")
		assert.Equal(t, b2.Name, "linux", "b2 Name mismatch")
		assert.Equal(t, b2.USN, 0, "b2 USN mismatch")
		assert.Equal(t, b2.Dirty, false, "b2 Dirty mismatch")

		assert.Equal(t, n1.UUID, "43827b9a-c2b0-4c06-a290-97991c896653", "n1 UUID mismatch")
		assert.Equal(t, n1.Body, "Booleans have toString()", "n1 Body mismatch")
		assert.Equal(t, n1.AddedOn, int64(1515199943), "n1 AddedOn mismatch")
		assert.Equal(t, n1.Deleted, false, "n1 Deleted mismatch")
		assert.Equal(t, n1.Dirty, false, "n1 Dirty mismatch")
		assert.Equal(t, n1.USN, 0, "n1 USN mismatch")
	})
}

func TestRemoveNote(t *testing.T) {
	testCases := []struct {
		yesFlag bool
	}{
		{
			yesFlag: false,
		},
		{
			yesFlag: true,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("--yes=%t", tc.yesFlag), func(t *testing.T) {
			// Setup
			db := database.InitTestDB(t, fmt.Sprintf("%s/%s", opts.NADDir, consts.NADDBFileName), nil)
			testutils.Setup2(t, db)

			// Execute
			if tc.yesFlag {
				testutils.RunNADCmd(t, opts, binaryName, "remove", "-y", "1")
			} else {
				testutils.WaitNADCmd(t, opts, testutils.UserConfirm, binaryName, "remove", "1")
			}
			defer testutils.RemoveDir(t, opts.HomeDir)

			// Test
			var noteCount, bookCount, jsNoteCount, linuxNoteCount int
			database.MustScan(t, "counting books", db.QueryRow("SELECT count(*) FROM books"), &bookCount)
			database.MustScan(t, "counting notes", db.QueryRow("SELECT count(*) FROM notes"), &noteCount)
			database.MustScan(t, "counting js notes", db.QueryRow("SELECT count(*) FROM notes WHERE book_uuid = ?", "js-book-uuid"), &jsNoteCount)
			database.MustScan(t, "counting linux notes", db.QueryRow("SELECT count(*) FROM notes WHERE book_uuid = ?", "linux-book-uuid"), &linuxNoteCount)

			assert.Equalf(t, bookCount, 2, "book count mismatch")
			assert.Equalf(t, noteCount, 3, "note count mismatch")
			assert.Equal(t, jsNoteCount, 2, "js book should have 2 notes")
			assert.Equal(t, linuxNoteCount, 1, "linux book book should have 1 note")

			var b1, b2 database.Book
			var n1, n2, n3 database.Note
			database.MustScan(t, "getting b1",
				db.QueryRow("SELECT name, deleted, usn FROM books WHERE uuid = ?", "js-book-uuid"),

				&b1.Name, &b1.Deleted, &b1.USN)
			database.MustScan(t, "getting b2",
				db.QueryRow("SELECT name, deleted, usn FROM books WHERE uuid = ?", "linux-book-uuid"),
				&b2.Name, &b2.Deleted, &b2.USN)
			database.MustScan(t, "getting n1",
				db.QueryRow("SELECT uuid, body, added_on, deleted, dirty, usn FROM notes WHERE book_uuid = ? AND uuid = ?", "js-book-uuid", "f0d0fbb7-31ff-45ae-9f0f-4e429c0c797f"),
				&n1.UUID, &n1.Body, &n1.AddedOn, &n1.Deleted, &n1.Dirty, &n1.USN)
			database.MustScan(t, "getting n2",
				db.QueryRow("SELECT uuid, body, added_on, deleted, dirty, usn FROM notes WHERE book_uuid = ? AND uuid = ?", "js-book-uuid", "43827b9a-c2b0-4c06-a290-97991c896653"),
				&n2.UUID, &n2.Body, &n2.AddedOn, &n2.Deleted, &n2.Dirty, &n2.USN)
			database.MustScan(t, "getting n3",
				db.QueryRow("SELECT uuid, body, added_on, deleted, dirty, usn FROM notes WHERE book_uuid = ? AND uuid = ?", "linux-book-uuid", "3e065d55-6d47-42f2-a6bf-f5844130b2d2"),
				&n3.UUID, &n3.Body, &n3.AddedOn, &n3.Deleted, &n3.Dirty, &n3.USN)

			assert.Equal(t, b1.Name, "js", "b1 name mismatch")
			assert.Equal(t, b1.Deleted, false, "b1 deleted mismatch")
			assert.Equal(t, b1.Dirty, false, "b1 Dirty mismatch")
			assert.Equal(t, b1.USN, 111, "b1 usn mismatch")

			assert.Equal(t, b2.Name, "linux", "b2 name mismatch")
			assert.Equal(t, b2.Deleted, false, "b2 deleted mismatch")
			assert.Equal(t, b2.Dirty, false, "b2 Dirty mismatch")
			assert.Equal(t, b2.USN, 122, "b2 usn mismatch")

			assert.Equal(t, n1.UUID, "f0d0fbb7-31ff-45ae-9f0f-4e429c0c797f", "n1 should have UUID")
			assert.Equal(t, n1.Body, "", "n1 body mismatch")
			assert.Equal(t, n1.Deleted, true, "n1 deleted mismatch")
			assert.Equal(t, n1.Dirty, true, "n1 Dirty mismatch")
			assert.Equal(t, n1.USN, 11, "n1 usn mismatch")

			assert.Equal(t, n2.UUID, "43827b9a-c2b0-4c06-a290-97991c896653", "n2 should have UUID")
			assert.Equal(t, n2.Body, "n2 body", "n2 body mismatch")
			assert.Equal(t, n2.Deleted, false, "n2 deleted mismatch")
			assert.Equal(t, n2.Dirty, false, "n2 Dirty mismatch")
			assert.Equal(t, n2.USN, 12, "n2 usn mismatch")

			assert.Equal(t, n3.UUID, "3e065d55-6d47-42f2-a6bf-f5844130b2d2", "n3 should have UUID")
			assert.Equal(t, n3.Body, "n3 body", "n3 body mismatch")
			assert.Equal(t, n3.Deleted, false, "n3 deleted mismatch")
			assert.Equal(t, n3.Dirty, false, "n3 Dirty mismatch")
			assert.Equal(t, n3.USN, 13, "n3 usn mismatch")
		})
	}
}

func TestRemoveBook(t *testing.T) {
	testCases := []struct {
		yesFlag bool
	}{
		{
			yesFlag: false,
		},
		{
			yesFlag: true,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("--yes=%t", tc.yesFlag), func(t *testing.T) {
			// Setup
			db := database.InitTestDB(t, fmt.Sprintf("%s/%s", opts.NADDir, consts.NADDBFileName), nil)
			testutils.Setup2(t, db)

			// Execute
			if tc.yesFlag {
				testutils.RunNADCmd(t, opts, binaryName, "remove", "-y", "js")
			} else {
				testutils.WaitNADCmd(t, opts, testutils.UserConfirm, binaryName, "remove", "js")
			}

			defer testutils.RemoveDir(t, opts.HomeDir)

			// Test
			var noteCount, bookCount, jsNoteCount, linuxNoteCount int
			database.MustScan(t, "counting books", db.QueryRow("SELECT count(*) FROM books"), &bookCount)
			database.MustScan(t, "counting notes", db.QueryRow("SELECT count(*) FROM notes"), &noteCount)
			database.MustScan(t, "counting js notes", db.QueryRow("SELECT count(*) FROM notes WHERE book_uuid = ?", "js-book-uuid"), &jsNoteCount)
			database.MustScan(t, "counting linux notes", db.QueryRow("SELECT count(*) FROM notes WHERE book_uuid = ?", "linux-book-uuid"), &linuxNoteCount)

			assert.Equalf(t, bookCount, 2, "book count mismatch")
			assert.Equalf(t, noteCount, 3, "note count mismatch")
			assert.Equal(t, jsNoteCount, 2, "js book should have 2 notes")
			assert.Equal(t, linuxNoteCount, 1, "linux book book should have 1 note")

			var b1, b2 database.Book
			var n1, n2, n3 database.Note
			database.MustScan(t, "getting b1",
				db.QueryRow("SELECT name, dirty, deleted, usn FROM books WHERE uuid = ?", "js-book-uuid"),

				&b1.Name, &b1.Dirty, &b1.Deleted, &b1.USN)
			database.MustScan(t, "getting b2",
				db.QueryRow("SELECT name, dirty, deleted, usn FROM books WHERE uuid = ?", "linux-book-uuid"),
				&b2.Name, &b2.Dirty, &b2.Deleted, &b2.USN)
			database.MustScan(t, "getting n1",
				db.QueryRow("SELECT uuid, body, added_on, dirty, deleted, usn FROM notes WHERE book_uuid = ? AND uuid = ?", "js-book-uuid", "f0d0fbb7-31ff-45ae-9f0f-4e429c0c797f"),
				&n1.UUID, &n1.Body, &n1.AddedOn, &n1.Deleted, &n1.Dirty, &n1.USN)
			database.MustScan(t, "getting n2",
				db.QueryRow("SELECT uuid, body, added_on, dirty, deleted, usn FROM notes WHERE book_uuid = ? AND uuid = ?", "js-book-uuid", "43827b9a-c2b0-4c06-a290-97991c896653"),
				&n2.UUID, &n2.Body, &n2.AddedOn, &n2.Deleted, &n2.Dirty, &n2.USN)
			database.MustScan(t, "getting n3",
				db.QueryRow("SELECT uuid, body, added_on, dirty, deleted, usn FROM notes WHERE book_uuid = ? AND uuid = ?", "linux-book-uuid", "3e065d55-6d47-42f2-a6bf-f5844130b2d2"),
				&n3.UUID, &n3.Body, &n3.AddedOn, &n3.Deleted, &n3.Dirty, &n3.USN)

			assert.NotEqual(t, b1.Name, "js", "b1 name mismatch")
			assert.Equal(t, b1.Dirty, true, "b1 Dirty mismatch")
			assert.Equal(t, b1.Deleted, true, "b1 deleted mismatch")
			assert.Equal(t, b1.USN, 111, "b1 usn mismatch")

			assert.Equal(t, b2.Name, "linux", "b2 name mismatch")
			assert.Equal(t, b2.Dirty, false, "b2 Dirty mismatch")
			assert.Equal(t, b2.Deleted, false, "b2 deleted mismatch")
			assert.Equal(t, b2.USN, 122, "b2 usn mismatch")

			assert.Equal(t, n1.UUID, "f0d0fbb7-31ff-45ae-9f0f-4e429c0c797f", "n1 should have UUID")
			assert.Equal(t, n1.Body, "", "n1 body mismatch")
			assert.Equal(t, n1.Dirty, true, "n1 Dirty mismatch")
			assert.Equal(t, n1.Deleted, true, "n1 deleted mismatch")
			assert.Equal(t, n1.USN, 11, "n1 usn mismatch")

			assert.Equal(t, n2.UUID, "43827b9a-c2b0-4c06-a290-97991c896653", "n2 should have UUID")
			assert.Equal(t, n2.Body, "", "n2 body mismatch")
			assert.Equal(t, n2.Dirty, true, "n2 Dirty mismatch")
			assert.Equal(t, n2.Deleted, true, "n2 deleted mismatch")
			assert.Equal(t, n2.USN, 12, "n2 usn mismatch")

			assert.Equal(t, n3.UUID, "3e065d55-6d47-42f2-a6bf-f5844130b2d2", "n3 should have UUID")
			assert.Equal(t, n3.Body, "n3 body", "n3 body mismatch")
			assert.Equal(t, n3.Dirty, false, "n3 Dirty mismatch")
			assert.Equal(t, n3.Deleted, false, "n3 deleted mismatch")
			assert.Equal(t, n3.USN, 13, "n3 usn mismatch")
		})
	}
}
