package controllers

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	"github.com/nadproject/nad/pkg/assert"
	"github.com/nadproject/nad/pkg/clock"
	"github.com/nadproject/nad/pkg/server/config"
	"github.com/nadproject/nad/pkg/server/models"
	// "github.com/nadproject/nad/pkg/server/presenters"
	// "github.com/pkg/errors"
)

func TestNotesV1Create(t *testing.T) {
	// Set up
	cfg := config.Load()
	cfg.SetPageTemplateDir(testPageDir)
	defer models.ClearTestData(t, models.TestServices.DB)

	user, _ := models.SetupUser(t, models.TestServices.User, models.TestServices.Session, "alice@example.com", "pass1234")
	models.MustExec(t, models.TestServices.DB.Model(&user).Update("max_usn", 101), "preparing user max_usn")

	// Test
	notesC := NewNotes(cfg, models.TestServices.Note, models.TestServices.User, clock.NewMock(), models.TestServices.DB)

	b1 := models.Book{
		UserID: user.ID,
		Name:   "js",
		USN:    58,
	}
	models.MustExec(t, models.TestServices.DB.Save(&b1), "preparing b1")

	dat := fmt.Sprintf(`{"book_uuid": "%s", "content": "note content"}`, b1.UUID)
	req := newReq(t, "POST", "/v1/api/notes", dat)
	w := httpDo(t, notesC.V1Create, req, &user)

	assert.Equal(t, w.Code, http.StatusCreated, "status code mismatch")

	var noteRecord models.Note
	var bookRecord models.Book
	var userRecord models.User
	var bookCount, noteCount int
	models.MustExec(t, models.TestServices.DB.Model(&models.Book{}).Count(&bookCount), "counting books")
	models.MustExec(t, models.TestServices.DB.Model(&models.Note{}).Count(&noteCount), "counting notes")
	models.MustExec(t, models.TestServices.DB.First(&noteRecord), "finding note")
	models.MustExec(t, models.TestServices.DB.Where("id = ?", b1.ID).First(&bookRecord), "finding book")
	models.MustExec(t, models.TestServices.DB.Where("id = ?", user.ID).First(&userRecord), "finding user record")

	assert.Equalf(t, bookCount, 1, "book count mismatch")
	assert.Equalf(t, noteCount, 1, "note count mismatch")

	assert.Equal(t, bookRecord.Name, b1.Name, "book name mismatch")
	assert.Equal(t, bookRecord.UUID, b1.UUID, "book uuid mismatch")
	assert.Equal(t, bookRecord.UserID, b1.UserID, "book user_id mismatch")
	assert.Equal(t, bookRecord.USN, 58, "book usn mismatch")

	assert.NotEqual(t, noteRecord.UUID, "", "note uuid should have been generated")
	assert.Equal(t, noteRecord.BookUUID, b1.UUID, "note book_uuid mismatch")
	assert.Equal(t, noteRecord.Body, "note content", "note content mismatch")
	assert.Equal(t, noteRecord.USN, 102, "note usn mismatch")
}

func TestNotesV1Update(t *testing.T) {
	updatedBody := "some updated content"

	b1UUID := "37868a8e-a844-4265-9a4f-0be598084733"
	b2UUID := "8f3bd424-6aa5-4ed5-910d-e5b38ab09f8c"

	testCases := []struct {
		payload              string
		noteUUID             string
		noteBookUUID         string
		noteBody             string
		notePublic           bool
		noteDeleted          bool
		expectedNoteBody     string
		expectedNoteBookName string
		expectedNoteBookUUID string
		expectedNotePublic   bool
	}{
		{
			payload: fmt.Sprintf(`{
				"content": "%s"
			}`, updatedBody),
			noteUUID:             "ab50aa32-b232-40d8-b10f-10a7f9134053",
			noteBookUUID:         b1UUID,
			notePublic:           false,
			noteBody:             "original content",
			noteDeleted:          false,
			expectedNoteBookUUID: b1UUID,
			expectedNoteBody:     "some updated content",
			expectedNoteBookName: "css",
			expectedNotePublic:   false,
		},
		{
			payload: fmt.Sprintf(`{
				"book_uuid": "%s"
			}`, b1UUID),
			noteUUID:             "ab50aa32-b232-40d8-b10f-10a7f9134053",
			noteBookUUID:         b1UUID,
			notePublic:           false,
			noteBody:             "original content",
			noteDeleted:          false,
			expectedNoteBookUUID: b1UUID,
			expectedNoteBody:     "original content",
			expectedNoteBookName: "css",
			expectedNotePublic:   false,
		},
		{
			payload: fmt.Sprintf(`{
				"book_uuid": "%s"
			}`, b2UUID),
			noteUUID:             "ab50aa32-b232-40d8-b10f-10a7f9134053",
			noteBookUUID:         b1UUID,
			notePublic:           false,
			noteBody:             "original content",
			noteDeleted:          false,
			expectedNoteBookUUID: b2UUID,
			expectedNoteBody:     "original content",
			expectedNoteBookName: "js",
			expectedNotePublic:   false,
		},
		{
			payload: fmt.Sprintf(`{
				"book_uuid": "%s",
				"content": "%s"
			}`, b2UUID, updatedBody),
			noteUUID:             "ab50aa32-b232-40d8-b10f-10a7f9134053",
			noteBookUUID:         b1UUID,
			notePublic:           false,
			noteBody:             "original content",
			noteDeleted:          false,
			expectedNoteBookUUID: b2UUID,
			expectedNoteBody:     "some updated content",
			expectedNoteBookName: "js",
			expectedNotePublic:   false,
		},
		{
			payload: fmt.Sprintf(`{
				"book_uuid": "%s",
				"content": "%s"
			}`, b1UUID, updatedBody),
			noteUUID:             "ab50aa32-b232-40d8-b10f-10a7f9134053",
			noteBookUUID:         b1UUID,
			notePublic:           false,
			noteBody:             "",
			noteDeleted:          true,
			expectedNoteBookUUID: b1UUID,
			expectedNoteBody:     updatedBody,
			expectedNoteBookName: "js",
			expectedNotePublic:   false,
		},
		{
			payload: fmt.Sprintf(`{
				"public": %t
			}`, true),
			noteUUID:             "ab50aa32-b232-40d8-b10f-10a7f9134053",
			noteBookUUID:         b1UUID,
			notePublic:           false,
			noteBody:             "original content",
			noteDeleted:          false,
			expectedNoteBookUUID: b1UUID,
			expectedNoteBody:     "original content",
			expectedNoteBookName: "css",
			expectedNotePublic:   true,
		},
		{
			payload: fmt.Sprintf(`{
				"public": %t
			}`, false),
			noteUUID:             "ab50aa32-b232-40d8-b10f-10a7f9134053",
			noteBookUUID:         b1UUID,
			notePublic:           true,
			noteBody:             "original content",
			noteDeleted:          false,
			expectedNoteBookUUID: b1UUID,
			expectedNoteBody:     "original content",
			expectedNoteBookName: "css",
			expectedNotePublic:   false,
		},
		{
			payload: fmt.Sprintf(`{
				"content": "%s",
				"public": %t
			}`, updatedBody, false),
			noteUUID:             "ab50aa32-b232-40d8-b10f-10a7f9134053",
			noteBookUUID:         b1UUID,
			notePublic:           true,
			noteBody:             "original content",
			noteDeleted:          false,
			expectedNoteBookUUID: b1UUID,
			expectedNoteBody:     updatedBody,
			expectedNoteBookName: "css",
			expectedNotePublic:   false,
		},
		{
			payload: fmt.Sprintf(`{
				"book_uuid": "%s",
				"content": "%s",
				"public": %t
			}`, b2UUID, updatedBody, true),
			noteUUID:             "ab50aa32-b232-40d8-b10f-10a7f9134053",
			noteBookUUID:         b1UUID,
			notePublic:           false,
			noteBody:             "original content",
			noteDeleted:          false,
			expectedNoteBookUUID: b2UUID,
			expectedNoteBody:     updatedBody,
			expectedNoteBookName: "js",
			expectedNotePublic:   true,
		},
	}

	for idx, tc := range testCases {
		t.Run(fmt.Sprintf("test case %d", idx), func(t *testing.T) {
			// Set up
			cfg := config.Load()
			cfg.SetPageTemplateDir(testPageDir)
			defer models.ClearTestData(t, models.TestServices.DB)

			user, _ := models.SetupUser(t, models.TestServices.User, models.TestServices.Session, "alice@example.com", "pass1234")
			models.MustExec(t, models.TestServices.DB.Model(&user).Update("max_usn", 101), "preparing user max_usn")

			b1 := models.Book{
				UUID:   b1UUID,
				UserID: user.ID,
				Name:   "css",
			}
			models.MustExec(t, models.TestServices.DB.Save(&b1), "preparing b1")
			b2 := models.Book{
				UUID:   b2UUID,
				UserID: user.ID,
				Name:   "js",
			}
			models.MustExec(t, models.TestServices.DB.Save(&b2), "preparing b2")

			note := models.Note{
				UserID:   user.ID,
				UUID:     tc.noteUUID,
				BookUUID: tc.noteBookUUID,
				Body:     tc.noteBody,
				Deleted:  tc.noteDeleted,
				Public:   tc.notePublic,
				AddedOn:  1579818739000000,
			}
			models.MustExec(t, models.TestServices.DB.Save(&note), "preparing note")

			// Execute
			notesC := NewNotes(cfg, models.TestServices.Note, models.TestServices.User, clock.NewMock(), models.TestServices.DB)
			endpoint := fmt.Sprintf("/v3/notes/%s", note.UUID)
			req := newReq(t, "PATCH", endpoint, tc.payload)
			req = mux.SetURLVars(req, map[string]string{"noteUUID": note.UUID})

			w := httpDo(t, notesC.V1Update, req, &user)

			// Test
			assert.Equal(t, w.Code, http.StatusOK, "status code mismatch")

			var bookRecord models.Book
			var noteRecord models.Note
			var userRecord models.User
			var noteCount, bookCount int
			models.MustExec(t, models.TestServices.DB.Model(&models.Book{}).Count(&bookCount), "counting books")
			models.MustExec(t, models.TestServices.DB.Model(&models.Note{}).Count(&noteCount), "counting notes")
			models.MustExec(t, models.TestServices.DB.Where("uuid = ?", note.UUID).First(&noteRecord), "finding note")
			models.MustExec(t, models.TestServices.DB.Where("id = ?", b1.ID).First(&bookRecord), "finding book")
			models.MustExec(t, models.TestServices.DB.Where("id = ?", user.ID).First(&userRecord), "finding user record")

			assert.Equalf(t, bookCount, 2, "book count mismatch")
			assert.Equalf(t, noteCount, 1, "note count mismatch")

			assert.Equal(t, noteRecord.UUID, tc.noteUUID, "note uuid mismatch for test case")
			assert.Equal(t, noteRecord.Body, tc.expectedNoteBody, "note content mismatch for test case")
			assert.Equal(t, noteRecord.BookUUID, tc.expectedNoteBookUUID, "note book_uuid mismatch for test case")
			// TODO: implement public notes
			// assert.Equal(t, noteRecord.Public, tc.expectedNotePublic, "note public mismatch for test case")
			assert.Equal(t, noteRecord.USN, 102, "note usn mismatch for test case")

			assert.Equal(t, userRecord.MaxUSN, 102, "user max_usn mismatch for test case")
		})
	}
}

func TestNotesV1Delete(t *testing.T) {
	b1UUID := "37868a8e-a844-4265-9a4f-0be598084733"

	testCases := []struct {
		content        string
		deleted        bool
		originalUSN    int
		expectedUSN    int
		expectedMaxUSN int
	}{
		{
			content:        "n1 content",
			deleted:        false,
			originalUSN:    12,
			expectedUSN:    982,
			expectedMaxUSN: 982,
		},
		{
			content:        "",
			deleted:        true,
			originalUSN:    12,
			expectedUSN:    982,
			expectedMaxUSN: 982,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("originally deleted %t", tc.deleted), func(t *testing.T) {
			// Set up
			cfg := config.Load()
			cfg.SetPageTemplateDir(testPageDir)
			defer models.ClearTestData(t, models.TestServices.DB)

			user, _ := models.SetupUser(t, models.TestServices.User, models.TestServices.Session, "alice@example.com", "pass1234")
			models.MustExec(t, models.TestServices.DB.Model(&user).Update("max_usn", 981), "preparing user max_usn")

			b1 := models.Book{
				UUID:   b1UUID,
				UserID: user.ID,
				Name:   "js",
			}
			models.MustExec(t, models.TestServices.DB.Save(&b1), "preparing b1")
			note := models.Note{
				UserID:   user.ID,
				BookUUID: b1.UUID,
				Body:     tc.content,
				Deleted:  tc.deleted,
				USN:      tc.originalUSN,
				AddedOn:  1579818739000000,
			}
			models.MustExec(t, models.TestServices.DB.Save(&note), "preparing note")

			// Execute
			notesC := NewNotes(cfg, models.TestServices.Note, models.TestServices.User, clock.NewMock(), models.TestServices.DB)

			endpoint := fmt.Sprintf("/api/v1/notes/%s", note.UUID)
			req := newReq(t, "POST", endpoint, "")
			req = mux.SetURLVars(req, map[string]string{"noteUUID": note.UUID})
			w := httpDo(t, notesC.V1Delete, req, &user)

			// Test
			assert.Equal(t, w.Code, http.StatusOK, "status code mismatch")

			var bookRecord models.Book
			var noteRecord models.Note
			var userRecord models.User
			var bookCount, noteCount int
			models.MustExec(t, models.TestServices.DB.Model(&models.Book{}).Count(&bookCount), "counting books")
			models.MustExec(t, models.TestServices.DB.Model(&models.Note{}).Count(&noteCount), "counting notes")
			models.MustExec(t, models.TestServices.DB.Where("uuid = ?", note.UUID).First(&noteRecord), "finding note")
			models.MustExec(t, models.TestServices.DB.Where("id = ?", b1.ID).First(&bookRecord), "finding book")
			models.MustExec(t, models.TestServices.DB.Where("id = ?", user.ID).First(&userRecord), "finding user record")

			assert.Equalf(t, bookCount, 1, "book count mismatch")
			assert.Equalf(t, noteCount, 1, "note count mismatch")

			assert.Equal(t, noteRecord.UUID, note.UUID, "note uuid mismatch for test case")
			assert.Equal(t, noteRecord.Body, "", "note content mismatch for test case")
			assert.Equal(t, noteRecord.Deleted, true, "note deleted mismatch for test case")
			assert.Equal(t, noteRecord.BookUUID, note.BookUUID, "note book_uuid mismatch for test case")
			assert.Equal(t, noteRecord.UserID, note.UserID, "note user_id mismatch for test case")
			assert.Equal(t, noteRecord.USN, tc.expectedUSN, "note usn mismatch for test case")

			assert.Equal(t, userRecord.MaxUSN, tc.expectedMaxUSN, "user max_usn mismatch for test case")
		})
	}
}
