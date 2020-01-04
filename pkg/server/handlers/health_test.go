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

package handlers

import (
	"net/http"
	"testing"

	"github.com/nadproject/nad/pkg/assert"
	"github.com/nadproject/nad/pkg/clock"
	"github.com/nadproject/nad/pkg/server/app"
	"github.com/nadproject/nad/pkg/server/testutils"
	"github.com/jinzhu/gorm"
)

func TestCheckHealth(t *testing.T) {
	// Setup
	server := MustNewServer(t, &app.App{
		DB:    &gorm.DB{},
		Clock: clock.NewMock(),
	})
	defer server.Close()

	// Execute
	req := testutils.MakeReq(server.URL, "GET", "/health", "")
	res := testutils.HTTPDo(t, req)

	// Test
	assert.StatusCodeEquals(t, res, http.StatusOK, "Status code mismtach")
}
