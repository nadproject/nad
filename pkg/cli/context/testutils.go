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

package context

import (
	"fmt"
	"testing"

	"github.com/nadproject/nad/pkg/cli/consts"
	"github.com/nadproject/nad/pkg/cli/database"
	"github.com/nadproject/nad/pkg/clock"
)

// InitTestCtx initializes a test context
func InitTestCtx(t *testing.T, nadDir string, dbOpts *database.TestDBOptions) NadCtx {
	dbPath := fmt.Sprintf("%s/%s", nadDir, consts.NADDBFileName)

	db := database.InitTestDB(t, dbPath, dbOpts)

	return NadCtx{
		DB:       db,
		NADDir: nadDir,
		// Use a mock clock to test times
		Clock: clock.NewMock(),
	}
}

// TeardownTestCtx cleans up the test context
func TeardownTestCtx(t *testing.T, ctx NadCtx) {
	database.CloseTestDB(t, ctx.DB)
}
