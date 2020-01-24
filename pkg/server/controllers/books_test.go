package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	"github.com/nadproject/nad/pkg/assert"
	"github.com/nadproject/nad/pkg/clock"
	"github.com/nadproject/nad/pkg/server/config"
	"github.com/nadproject/nad/pkg/server/models"
	"github.com/nadproject/nad/pkg/server/presenters"
	"github.com/pkg/errors"
)

func TestBooksV1Create(t *testing.T) {
	// Set up
	cfg := config.Load()
	cfg.SetPageTemplateDir(testPageDir)
	defer models.ClearTestData(t, models.TestServices.DB)

	user, _ := models.SetupUser(t, models.TestServices.User, models.TestServices.Session, "alice@example.com", "pass1234")
	models.MustExec(t, models.TestServices.DB.Model(&user).Update("max_usn", 101), "preparing user max_usn")

	// Test
	booksC := NewBooks(cfg, models.TestServices.Book, models.TestServices.User, models.TestServices.Note, clock.NewMock(), models.TestServices.DB)
	req := newReq(t, "POST", "/v1/api/books", `{"name": "js"}`)
	w := httpDo(t, booksC.V1Create, req, &user)
	assert.Equal(t, w.Code, http.StatusCreated, "status code mismatch")

	var bookRecord models.Book
	var userRecord models.User
	var bookCount, noteCount int
	models.MustExec(t, models.TestServices.DB.Model(models.Book{}).Count(&bookCount), "counting books")
	models.MustExec(t, models.TestServices.DB.Model(models.Note{}).Count(&noteCount), "counting notes")
	models.MustExec(t, models.TestServices.DB.First(&bookRecord), "finding book")
	models.MustExec(t, models.TestServices.DB.Where("id = ?", user.ID).First(&userRecord), "finding user record")

	maxUSN := 102

	assert.Equalf(t, bookCount, 1, "book count mismatch")
	assert.Equalf(t, noteCount, 0, "note count mismatch")

	assert.NotEqual(t, bookRecord.UUID, "", "book uuid should have been generated")
	assert.Equal(t, bookRecord.Name, "js", "book name mismatch")
	assert.Equal(t, bookRecord.UserID, user.ID, "book user_id mismatch")
	assert.Equal(t, bookRecord.USN, maxUSN, "book user_id mismatch")
	assert.Equal(t, userRecord.MaxUSN, maxUSN, "user max_usn mismatch")

	var got presenters.Book
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatal(errors.Wrap(err, "decoding got"))
	}
	expected := presenters.Book{
		UUID:      bookRecord.UUID,
		USN:       bookRecord.USN,
		CreatedAt: bookRecord.CreatedAt,
		UpdatedAt: bookRecord.UpdatedAt,
		Name:      "js",
	}

	assert.DeepEqual(t, got, expected, "payload mismatch")
}

func TestBooksV1CreateDuplicate(t *testing.T) {
	// Set up
	cfg := config.Load()
	cfg.SetPageTemplateDir(testPageDir)
	defer models.ClearTestData(t, models.TestServices.DB)

	user, _ := models.SetupUser(t, models.TestServices.User, models.TestServices.Session, "alice@example.com", "pass1234")
	models.MustExec(t, models.TestServices.DB.Model(&user).Update("max_usn", 101), "preparing user max_usn")

	b1 := models.Book{
		UserID: user.ID,
		Name:   "js",
		USN:    58,
	}
	models.MustExec(t, models.TestServices.DB.Save(&b1), "preparing book data")

	// Test
	booksC := NewBooks(cfg, models.TestServices.Book, models.TestServices.User, models.TestServices.Note, clock.NewMock(), models.TestServices.DB)
	req := newReq(t, "POST", "/v1/api/books", `{"name": "js"}`)
	w := httpDo(t, booksC.V1Create, req, &user)
	assert.Equal(t, w.Code, http.StatusConflict, "status code mismatch")

	var bookRecord models.Book
	var userRecord models.User
	var bookCount, noteCount int
	models.MustExec(t, models.TestServices.DB.Model(models.Book{}).Count(&bookCount), "counting books")
	models.MustExec(t, models.TestServices.DB.Model(models.Note{}).Count(&noteCount), "counting notes")
	models.MustExec(t, models.TestServices.DB.First(&bookRecord), "finding book")
	models.MustExec(t, models.TestServices.DB.Where("id = ?", user.ID).First(&userRecord), "finding user record")

	assert.Equalf(t, bookCount, 1, "book count mismatch")
	assert.Equalf(t, noteCount, 0, "note count mismatch")

	assert.NotEqual(t, bookRecord.UUID, "", "book uuid should have been generated")
	assert.Equal(t, bookRecord.Name, "js", "book name mismatch")
	assert.Equal(t, bookRecord.UserID, user.ID, "book user_id mismatch")
	assert.Equal(t, bookRecord.USN, b1.USN, "book user_id mismatch")
	assert.Equal(t, userRecord.MaxUSN, 101, "user max_usn mismatch")
}

func TestBooksV1Delete(t *testing.T) {
	testCases := []struct {
		label          string
		deleted        bool
		expectedB2USN  int
		expectedMaxUSN int
		expectedN2USN  int
		expectedN3USN  int
	}{
		{
			label:          "n1 content",
			deleted:        false,
			expectedMaxUSN: 61,
			expectedB2USN:  61,
			expectedN2USN:  59,
			expectedN3USN:  60,
		},
		{
			label:          "",
			deleted:        true,
			expectedMaxUSN: 59,
			expectedB2USN:  59,
			expectedN2USN:  5,
			expectedN3USN:  6,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("originally deleted %t", tc.deleted), func(t *testing.T) {
			// Set up
			cfg := config.Load()
			cfg.SetPageTemplateDir(testPageDir)
			defer models.ClearTestData(t, models.TestServices.DB)

			user, _ := models.SetupUser(t, models.TestServices.User, models.TestServices.Session, "alice@example.com", "pass1234")
			models.MustExec(t, models.TestServices.DB.Model(&user).Update("max_usn", 58), "preparing user max_usn")

			anotherUser, _ := models.SetupUser(t, models.TestServices.User, models.TestServices.Session, "bob@example.com", "pass1234")
			models.MustExec(t, models.TestServices.DB.Model(&anotherUser).Update("max_usn", 109), "preparing user max_usn")

			b1 := models.Book{
				UserID:  user.ID,
				Name:    "js",
				USN:     1,
				AddedOn: 1579750817,
			}
			models.MustExec(t, models.TestServices.DB.Save(&b1), "preparing book data")
			b2 := models.Book{
				UserID:  user.ID,
				Name:    tc.label,
				USN:     2,
				Deleted: tc.deleted,
				AddedOn: 1579750817,
			}
			models.MustExec(t, models.TestServices.DB.Save(&b2), "preparing book data")
			b3 := models.Book{
				UserID:  anotherUser.ID,
				Name:    "linux",
				USN:     3,
				AddedOn: 1579750817,
			}
			models.MustExec(t, models.TestServices.DB.Save(&b3), "preparing book data")

			var n2Body string
			if !tc.deleted {
				n2Body = "n2 content"
			}
			var n3Body string
			if !tc.deleted {
				n3Body = "n3 content"
			}

			n1 := models.Note{
				UserID:   user.ID,
				BookUUID: b1.UUID,
				Body:     "n1 content",
				USN:      4,
				AddedOn:  1579750817,
			}
			models.MustExec(t, models.TestServices.DB.Save(&n1), "preparing book data")
			n2 := models.Note{
				UserID:   user.ID,
				BookUUID: b2.UUID,
				Body:     n2Body,
				USN:      5,
				Deleted:  tc.deleted,
				AddedOn:  1579750817,
			}
			models.MustExec(t, models.TestServices.DB.Save(&n2), "preparing book data")
			n3 := models.Note{
				UserID:   user.ID,
				BookUUID: b2.UUID,
				Body:     n3Body,
				USN:      6,
				Deleted:  tc.deleted,
				AddedOn:  1579750817,
			}
			models.MustExec(t, models.TestServices.DB.Save(&n3), "preparing book data")
			n4 := models.Note{
				UserID:   user.ID,
				BookUUID: b2.UUID,
				Body:     "",
				USN:      7,
				Deleted:  true,
				AddedOn:  1579750817,
			}
			models.MustExec(t, models.TestServices.DB.Save(&n4), "preparing book data")
			n5 := models.Note{
				UserID:   anotherUser.ID,
				BookUUID: b3.UUID,
				Body:     "n5 content",
				USN:      8,
				AddedOn:  1579750817,
			}
			models.MustExec(t, models.TestServices.DB.Save(&n5), "preparing book data")

			booksC := NewBooks(cfg, models.TestServices.Book, models.TestServices.User, models.TestServices.Note, clock.NewMock(), models.TestServices.DB)
			req := newReq(t, "DELETE", fmt.Sprintf("/v1/api/books/%s", b2.UUID), "")
			req = mux.SetURLVars(req, map[string]string{"bookUUID": b2.UUID})
			w := httpDo(t, booksC.V1Delete, req, &user)

			// Test
			assert.Equal(t, w.Code, http.StatusOK, "")

			var b1Record, b2Record, b3Record models.Book
			var n1Record, n2Record, n3Record, n4Record, n5Record models.Note
			var userRecord models.User
			var bookCount, noteCount int

			models.MustExec(t, models.TestServices.DB.Model(&models.Book{}).Count(&bookCount), "counting books")
			models.MustExec(t, models.TestServices.DB.Model(&models.Note{}).Count(&noteCount), "counting notes")
			models.MustExec(t, models.TestServices.DB.Where("id = ?", b1.ID).First(&b1Record), "finding b1")
			models.MustExec(t, models.TestServices.DB.Where("id = ?", b2.ID).First(&b2Record), "finding b2")
			models.MustExec(t, models.TestServices.DB.Where("id = ?", b3.ID).First(&b3Record), "finding b3")
			models.MustExec(t, models.TestServices.DB.Where("id = ?", n1.ID).First(&n1Record), "finding n1")
			models.MustExec(t, models.TestServices.DB.Where("id = ?", n2.ID).First(&n2Record), "finding n2")
			models.MustExec(t, models.TestServices.DB.Where("id = ?", n3.ID).First(&n3Record), "finding n3")
			models.MustExec(t, models.TestServices.DB.Where("id = ?", n4.ID).First(&n4Record), "finding n4")
			models.MustExec(t, models.TestServices.DB.Where("id = ?", n5.ID).First(&n5Record), "finding n5")
			models.MustExec(t, models.TestServices.DB.Where("id = ?", user.ID).First(&userRecord), "finding user record")

			assert.Equal(t, bookCount, 3, "book count mismatch")
			assert.Equal(t, noteCount, 5, "note count mismatch")

			assert.Equal(t, userRecord.MaxUSN, tc.expectedMaxUSN, "user max_usn mismatch")

			assert.Equal(t, b1Record.Deleted, false, "b1 deleted mismatch")
			assert.Equal(t, b1Record.Name, b1.Name, "b1 content mismatch")
			assert.Equal(t, b1Record.USN, b1.USN, "b1 usn mismatch")
			assert.Equal(t, b2Record.Deleted, true, "b2 deleted mismatch")
			assert.Equal(t, b2Record.Name, "", "b2 content mismatch")
			assert.Equal(t, b2Record.USN, tc.expectedB2USN, "b2 usn mismatch")
			assert.Equal(t, b3Record.Deleted, false, "b3 deleted mismatch")
			assert.Equal(t, b3Record.Name, b3.Name, "b3 content mismatch")
			assert.Equal(t, b3Record.USN, b3.USN, "b3 usn mismatch")

			assert.Equal(t, n1Record.USN, n1.USN, "n1 usn mismatch")
			assert.Equal(t, n1Record.Deleted, false, "n1 deleted mismatch")
			assert.Equal(t, n1Record.Body, n1.Body, "n1 content mismatch")

			assert.Equal(t, n2Record.USN, tc.expectedN2USN, "n2 usn mismatch")
			assert.Equal(t, n2Record.Deleted, true, "n2 deleted mismatch")
			assert.Equal(t, n2Record.Body, "", "n2 content mismatch")

			assert.Equal(t, n3Record.USN, tc.expectedN3USN, "n3 usn mismatch")
			assert.Equal(t, n3Record.Deleted, true, "n3 deleted mismatch")
			assert.Equal(t, n3Record.Body, "", "n3 content mismatch")

			// if already deleted, usn should remain the same and hence should not contribute to bumping the max_usn
			assert.Equal(t, n4Record.USN, n4.USN, "n4 usn mismatch")
			assert.Equal(t, n4Record.Deleted, true, "n4 deleted mismatch")
			assert.Equal(t, n4Record.Body, "", "n4 content mismatch")

			assert.Equal(t, n5Record.USN, n5.USN, "n5 usn mismatch")
			assert.Equal(t, n5Record.Deleted, false, "n5 deleted mismatch")
			assert.Equal(t, n5Record.Body, n5.Body, "n5 content mismatch")
		})
	}
}

func TestBooksV1Update(t *testing.T) {
	updatedLabel := "updated-label"

	b1UUID := "ead8790f-aff9-4bdf-8eec-f734ccd29202"
	b2UUID := "0ecaac96-8d72-4e04-8925-5a21b79a16da"

	testCases := []struct {
		payload           string
		bookUUID          string
		bookDeleted       bool
		bookLabel         string
		expectedBookLabel string
	}{
		{
			payload: fmt.Sprintf(`{
				"name": "%s"
			}`, updatedLabel),
			bookUUID:          b1UUID,
			bookDeleted:       false,
			bookLabel:         "original-label",
			expectedBookLabel: updatedLabel,
		},
		// if a deleted book is updated, it should be un-deleted
		{
			payload: fmt.Sprintf(`{
				"name": "%s"
			}`, updatedLabel),
			bookUUID:          b1UUID,
			bookDeleted:       true,
			bookLabel:         "",
			expectedBookLabel: updatedLabel,
		},
	}

	for idx, tc := range testCases {
		t.Run(fmt.Sprintf("test case %d", idx), func(t *testing.T) {
			cfg := config.Load()
			cfg.SetPageTemplateDir(testPageDir)
			defer models.ClearTestData(t, models.TestServices.DB)

			// Setup
			user, _ := models.SetupUser(t, models.TestServices.User, models.TestServices.Session, "alice@example.com", "pass1234")
			models.MustExec(t, models.TestServices.DB.Model(&user).Update("max_usn", 101), "preparing user max_usn")

			b1 := models.Book{
				UUID:    tc.bookUUID,
				UserID:  user.ID,
				Name:    tc.bookLabel,
				Deleted: tc.bookDeleted,
			}
			models.MustExec(t, models.TestServices.DB.Save(&b1), "preparing b1")
			b2 := models.Book{
				UUID:   b2UUID,
				UserID: user.ID,
				Name:   "js",
			}
			models.MustExec(t, models.TestServices.DB.Save(&b2), "preparing b2")

			// Executdb,e
			booksC := NewBooks(cfg, models.TestServices.Book, models.TestServices.User, models.TestServices.Note, clock.NewMock(), models.TestServices.DB)
			req := newReq(t, "PATCH", fmt.Sprintf("/v1/api/books/%s", b2.UUID), tc.payload)
			req = mux.SetURLVars(req, map[string]string{"bookUUID": tc.bookUUID})
			w := httpDo(t, booksC.V1Update, req, &user)

			// Test
			assert.Equal(t, w.Code, http.StatusOK, "status code mismatch for test case")

			var bookRecord models.Book
			var userRecord models.User
			var noteCount, bookCount int
			models.MustExec(t, models.TestServices.DB.Model(&models.Book{}).Count(&bookCount), "counting books")
			models.MustExec(t, models.TestServices.DB.Model(&models.Note{}).Count(&noteCount), "counting notes")
			models.MustExec(t, models.TestServices.DB.Where("id = ?", b1.ID).First(&bookRecord), "finding book")
			models.MustExec(t, models.TestServices.DB.Where("id = ?", user.ID).First(&userRecord), "finding user record")

			assert.Equalf(t, bookCount, 2, "book count mismatch")
			assert.Equalf(t, noteCount, 0, "note count mismatch")

			assert.Equalf(t, bookRecord.UUID, tc.bookUUID, "book uuid mismatch")
			assert.Equalf(t, bookRecord.Name, tc.expectedBookLabel, "book label mismatch")
			assert.Equalf(t, bookRecord.USN, 102, "book usn mismatch")
			assert.Equalf(t, bookRecord.Deleted, false, "book Deleted mismatch")

			assert.Equal(t, userRecord.MaxUSN, 102, fmt.Sprintf("user max_usn mismatch for test case %d", idx))
		})
	}
}

func TestBooksV1Get(t *testing.T) {
	// Set up
	cfg := config.Load()
	cfg.SetPageTemplateDir(testPageDir)
	defer models.ClearTestData(t, models.TestServices.DB)

	user, _ := models.SetupUser(t, models.TestServices.User, models.TestServices.Session, "alice@example.com", "pass1234")
	anotherUser, _ := models.SetupUser(t, models.TestServices.User, models.TestServices.Session, "bob@example.com", "pass1234")

	b1 := models.Book{
		UserID:  user.ID,
		Name:    "js",
		USN:     1123,
		Deleted: false,
	}
	models.MustExec(t, models.TestServices.DB.Save(&b1), "preparing b1")
	b2 := models.Book{
		UserID:  user.ID,
		Name:    "css",
		USN:     1125,
		Deleted: false,
	}
	models.MustExec(t, models.TestServices.DB.Save(&b2), "preparing b2")
	b3 := models.Book{
		UserID:  anotherUser.ID,
		Name:    "css",
		USN:     1128,
		Deleted: false,
	}
	models.MustExec(t, models.TestServices.DB.Save(&b3), "preparing b3")
	b4 := models.Book{
		UserID:  user.ID,
		Name:    "",
		USN:     1129,
		Deleted: true,
	}
	models.MustExec(t, models.TestServices.DB.Save(&b4), "preparing b4")

	// Execute
	req := newReq(t, "GET", fmt.Sprintf("/v1/api/books/%s", b2.UUID), "")
	booksC := NewBooks(cfg, models.TestServices.Book, models.TestServices.User, models.TestServices.Note, clock.NewMock(), models.TestServices.DB)
	w := httpDo(t, booksC.V1Index, req, &user)

	// Test
	assert.Equal(t, w.Code, http.StatusOK, "status code mismatch")

	var payload []presenters.Book
	if err := json.NewDecoder(w.Body).Decode(&payload); err != nil {
		t.Fatal(errors.Wrap(err, "decoding payload"))
	}

	var b1Record, b2Record models.Book
	models.MustExec(t, models.TestServices.DB.Where("id = ?", b1.ID).First(&b1Record), "finding b1")
	models.MustExec(t, models.TestServices.DB.Where("id = ?", b2.ID).First(&b2Record), "finding b2")
	models.MustExec(t, models.TestServices.DB.Where("id = ?", b2.ID).First(&b2Record), "finding b2")

	expected := []presenters.Book{
		{
			UUID:      b2Record.UUID,
			CreatedAt: b2Record.CreatedAt,
			UpdatedAt: b2Record.UpdatedAt,
			Name:      b2Record.Name,
			USN:       b2Record.USN,
		},
		{
			UUID:      b1Record.UUID,
			CreatedAt: b1Record.CreatedAt,
			UpdatedAt: b1Record.UpdatedAt,
			Name:      b1Record.Name,
			USN:       b1Record.USN,
		},
	}

	assert.DeepEqual(t, payload, expected, "payload mismatch")
}

func TestBooksV1Get_Name(t *testing.T) {
	// Set up
	cfg := config.Load()
	cfg.SetPageTemplateDir(testPageDir)
	defer models.ClearTestData(t, models.TestServices.DB)

	user, _ := models.SetupUser(t, models.TestServices.User, models.TestServices.Session, "alice@example.com", "pass1234")
	anotherUser, _ := models.SetupUser(t, models.TestServices.User, models.TestServices.Session, "bob@example.com", "pass1234")

	b1 := models.Book{
		UserID: user.ID,
		Name:   "js",
	}
	models.MustExec(t, models.TestServices.DB.Save(&b1), "preparing b1")
	b2 := models.Book{
		UserID: user.ID,
		Name:   "css",
	}
	models.MustExec(t, models.TestServices.DB.Save(&b2), "preparing b2")
	b3 := models.Book{
		UserID: anotherUser.ID,
		Name:   "js",
	}
	models.MustExec(t, models.TestServices.DB.Save(&b3), "preparing b3")

	// Execute
	req := newReq(t, "GET", "/api/v1/books?name=js", "")
	booksC := NewBooks(cfg, models.TestServices.Book, models.TestServices.User, models.TestServices.Note, clock.NewMock(), models.TestServices.DB)
	w := httpDo(t, booksC.V1Index, req, &user)

	// Test
	assert.Equal(t, w.Code, http.StatusOK, "status coe mismatmch")

	var payload []presenters.Book
	if err := json.NewDecoder(w.Body).Decode(&payload); err != nil {
		t.Fatal(errors.Wrap(err, "decoding payload"))
	}

	var b1Record models.Book
	models.MustExec(t, models.TestServices.DB.Where("id = ?", b1.ID).First(&b1Record), "finding b1")

	expected := []presenters.Book{
		{
			UUID:      b1Record.UUID,
			CreatedAt: b1Record.CreatedAt,
			UpdatedAt: b1Record.UpdatedAt,
			Name:      b1Record.Name,
			USN:       b1Record.USN,
		},
	}

	assert.DeepEqual(t, payload, expected, "payload mismatch")
}
