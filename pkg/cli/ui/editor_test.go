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

package ui

import (
	"fmt"
	"os"
	"testing"

	"github.com/nadproject/nad/pkg/assert"
	"github.com/nadproject/nad/pkg/cli/context"
	"github.com/pkg/errors"
)

func TestGetTmpContentPath(t *testing.T) {
	t.Run("no collision", func(t *testing.T) {
		ctx := context.InitTestCtx(t, "../tmp1", nil)
		defer context.TeardownTestCtx(t, ctx)

		res, err := GetTmpContentPath(ctx)
		if err != nil {
			t.Fatal(errors.Wrap(err, "executing"))
		}

		expected := fmt.Sprintf("%s/%s", ctx.NADDir, "NAD_TMPCONTENT_0.md")
		assert.Equal(t, res, expected, "filename did not match")
	})

	t.Run("one existing session", func(t *testing.T) {
		// set up
		ctx := context.InitTestCtx(t, "../tmp2", nil)
		defer context.TeardownTestCtx(t, ctx)

		p := fmt.Sprintf("%s/%s", ctx.NADDir, "NAD_TMPCONTENT_0.md")
		if _, err := os.Create(p); err != nil {
			t.Fatal(errors.Wrap(err, "preparing the conflicting file"))
		}

		// execute
		res, err := GetTmpContentPath(ctx)
		if err != nil {
			t.Fatal(errors.Wrap(err, "executing"))
		}

		// test
		expected := fmt.Sprintf("%s/%s", ctx.NADDir, "NAD_TMPCONTENT_1.md")
		assert.Equal(t, res, expected, "filename did not match")
	})

	t.Run("two existing sessions", func(t *testing.T) {
		// set up
		ctx := context.InitTestCtx(t, "../tmp3", nil)
		defer context.TeardownTestCtx(t, ctx)

		p1 := fmt.Sprintf("%s/%s", ctx.NADDir, "NAD_TMPCONTENT_0.md")
		if _, err := os.Create(p1); err != nil {
			t.Fatal(errors.Wrap(err, "preparing the conflicting file"))
		}
		p2 := fmt.Sprintf("%s/%s", ctx.NADDir, "NAD_TMPCONTENT_1.md")
		if _, err := os.Create(p2); err != nil {
			t.Fatal(errors.Wrap(err, "preparing the conflicting file"))
		}

		// execute
		res, err := GetTmpContentPath(ctx)
		if err != nil {
			t.Fatal(errors.Wrap(err, "executing"))
		}

		// test
		expected := fmt.Sprintf("%s/%s", ctx.NADDir, "NAD_TMPCONTENT_2.md")
		assert.Equal(t, res, expected, "filename did not match")
	})
}
