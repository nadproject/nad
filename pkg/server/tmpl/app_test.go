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

package tmpl

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/nadproject/nad/pkg/assert"
	"github.com/nadproject/nad/pkg/server/database"
	"github.com/nadproject/nad/pkg/server/testutils"
	"github.com/pkg/errors"
)

func TestAppShellExecute(t *testing.T) {
	t.Run("home", func(t *testing.T) {
		a, err := NewAppShell(testutils.DB, []byte("<head><title>{{ .Title }}</title>{{ .MetaTags }}</head>"))
		if err != nil {
			t.Fatal(errors.Wrap(err, "preparing app shell"))
		}

		r, err := http.NewRequest("GET", "http://mock.url/", nil)
		if err != nil {
			t.Fatal(errors.Wrap(err, "preparing request"))
		}

		b, err := a.Execute(r)
		if err != nil {
			t.Fatal(errors.Wrap(err, "executing"))
		}

		assert.Equal(t, string(b), "<head><title>nad</title></head>", "result mismatch")
	})

	t.Run("note", func(t *testing.T) {
		defer testutils.ClearData()

		user := testutils.SetupUserData()
		b1 := database.Book{
			UserID: user.ID,
			Label:  "js",
		}
		testutils.MustExec(t, testutils.DB.Save(&b1), "preparing b1")
		n1 := database.Note{
			UserID:   user.ID,
			BookUUID: b1.UUID,
			Public:   true,
			Body:     "n1 content",
		}
		testutils.MustExec(t, testutils.DB.Save(&n1), "preparing note")

		a, err := NewAppShell(testutils.DB, []byte("{{ .MetaTags }}"))
		if err != nil {
			t.Fatal(errors.Wrap(err, "preparing app shell"))
		}

		endpoint := fmt.Sprintf("http://mock.url/notes/%s", n1.UUID)
		r, err := http.NewRequest("GET", endpoint, nil)
		if err != nil {
			t.Fatal(errors.Wrap(err, "preparing request"))
		}

		b, err := a.Execute(r)
		if err != nil {
			t.Fatal(errors.Wrap(err, "executing"))
		}

		assert.NotEqual(t, string(b), "", "result should not be empty")
	})
}
