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
	"html/template"
	"testing"
	"time"

	"github.com/nadproject/nad/pkg/assert"
	"github.com/nadproject/nad/pkg/server/database"
	"github.com/nadproject/nad/pkg/server/testutils"
	"github.com/pkg/errors"
)

func TestDefaultPageGetData(t *testing.T) {
	p := defaultPage{}

	result := p.getData()

	assert.Equal(t, result.MetaTags, template.HTML(""), "MetaTags mismatch")
	assert.Equal(t, result.Title, "nad", "Title mismatch")
}

func TestNotePageGetData(t *testing.T) {
	a, err := NewAppShell(testutils.DB, nil)
	if err != nil {
		t.Fatal(errors.Wrap(err, "preparing app shell"))
	}

	p := notePage{
		Note: database.Note{
			Book: database.Book{
				Label: "vocabulary",
			},
			AddedOn: time.Date(2019, time.January, 2, 0, 0, 0, 0, time.UTC).UnixNano(),
		},
		T: a.T,
	}

	result, err := p.getData()
	if err != nil {
		t.Fatal(errors.Wrap(err, "executing"))
	}

	assert.NotEqual(t, result.MetaTags, template.HTML(""), "MetaTags should not be empty")
	assert.Equal(t, result.Title, "Note: vocabulary (Jan 2 2019)", "Title mismatch")
}
