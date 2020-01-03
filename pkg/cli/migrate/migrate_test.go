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
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nadproject/nad/pkg/assert"
	"github.com/nadproject/nad/pkg/cli/consts"
	"github.com/nadproject/nad/pkg/cli/context"
	"github.com/nadproject/nad/pkg/cli/database"
	"github.com/nadproject/nad/pkg/cli/testutils"
	"github.com/nadproject/nad/pkg/cli/utils"
	"github.com/pkg/errors"
)

func TestExecute_bump_schema(t *testing.T) {
	testCases := []struct {
		schemaKey string
	}{
		{
			schemaKey: consts.SystemSchema,
		},
		{
			schemaKey: consts.SystemRemoteSchema,
		},
	}

	for _, tc := range testCases {
		func() {
			// set up
			opts := database.TestDBOptions{SkipMigration: true}
			ctx := context.InitTestCtx(t, "../tmp", &opts)
			defer context.TeardownTestCtx(t, ctx)

			db := ctx.DB

			database.MustExec(t, "inserting a schema", db, "INSERT INTO system (key, value) VALUES (?, ?)", tc.schemaKey, 8)

			m1 := migration{
				name: "noop",
				run: func(ctx context.NadCtx, db *database.DB) error {
					return nil
				},
			}
			m2 := migration{
				name: "noop",
				run: func(ctx context.NadCtx, db *database.DB) error {
					return nil
				},
			}

			// execute
			err := execute(ctx, m1, tc.schemaKey)
			if err != nil {
				t.Fatal(errors.Wrap(err, "failed to execute"))
			}
			err = execute(ctx, m2, tc.schemaKey)
			if err != nil {
				t.Fatal(errors.Wrap(err, "failed to execute"))
			}

			// test
			var schema int
			database.MustScan(t, "getting schema", db.QueryRow("SELECT value FROM system WHERE key = ?", tc.schemaKey), &schema)
			assert.Equal(t, schema, 10, "schema was not incremented properly")
		}()
	}
}

func TestRun_nonfresh(t *testing.T) {
	testCases := []struct {
		mode      int
		schemaKey string
	}{
		{
			mode:      LocalMode,
			schemaKey: consts.SystemSchema,
		},
		{
			mode:      RemoteMode,
			schemaKey: consts.SystemRemoteSchema,
		},
	}

	for _, tc := range testCases {
		func() {
			// set up
			opts := database.TestDBOptions{SkipMigration: true}
			ctx := context.InitTestCtx(t, "../tmp", &opts)
			defer context.TeardownTestCtx(t, ctx)

			db := ctx.DB
			database.MustExec(t, "inserting a schema", db, "INSERT INTO system (key, value) VALUES (?, ?)", tc.schemaKey, 2)
			database.MustExec(t, "creating a temporary table for testing", db,
				"CREATE TABLE migrate_run_test ( name string )")

			sequence := []migration{
				{
					name: "v1",
					run: func(ctx context.NadCtx, db *database.DB) error {
						database.MustExec(t, "marking v1 completed", db, "INSERT INTO migrate_run_test (name) VALUES (?)", "v1")
						return nil
					},
				},
				{
					name: "v2",
					run: func(ctx context.NadCtx, db *database.DB) error {
						database.MustExec(t, "marking v2 completed", db, "INSERT INTO migrate_run_test (name) VALUES (?)", "v2")
						return nil
					},
				},
				{
					name: "v3",
					run: func(ctx context.NadCtx, db *database.DB) error {
						database.MustExec(t, "marking v3 completed", db, "INSERT INTO migrate_run_test (name) VALUES (?)", "v3")
						return nil
					},
				},
				{
					name: "v4",
					run: func(ctx context.NadCtx, db *database.DB) error {
						database.MustExec(t, "marking v4 completed", db, "INSERT INTO migrate_run_test (name) VALUES (?)", "v4")
						return nil
					},
				},
			}

			// execute
			err := Run(ctx, sequence, tc.mode)
			if err != nil {
				t.Fatal(errors.Wrap(err, "failed to run"))
			}

			// test
			var schema int
			database.MustScan(t, fmt.Sprintf("getting schema for %s", tc.schemaKey), db.QueryRow("SELECT value FROM system WHERE key = ?", tc.schemaKey), &schema)
			assert.Equal(t, schema, 4, fmt.Sprintf("schema was not updated for %s", tc.schemaKey))

			var testRunCount int
			database.MustScan(t, "counting test runs", db.QueryRow("SELECT count(*) FROM migrate_run_test"), &testRunCount)
			assert.Equal(t, testRunCount, 2, "test run count mismatch")

			var testRun1, testRun2 string
			database.MustScan(t, "finding test run 1", db.QueryRow("SELECT name FROM migrate_run_test WHERE name = ?", "v3"), &testRun1)
			database.MustScan(t, "finding test run 2", db.QueryRow("SELECT name FROM migrate_run_test WHERE name = ?", "v4"), &testRun2)
		}()
	}
}

func TestRun_fresh(t *testing.T) {
	testCases := []struct {
		mode      int
		schemaKey string
	}{
		{
			mode:      LocalMode,
			schemaKey: consts.SystemSchema,
		},
		{
			mode:      RemoteMode,
			schemaKey: consts.SystemRemoteSchema,
		},
	}

	for _, tc := range testCases {
		func() {
			// set up
			opts := database.TestDBOptions{SkipMigration: true}
			ctx := context.InitTestCtx(t, "../tmp", &opts)
			defer context.TeardownTestCtx(t, ctx)

			db := ctx.DB

			database.MustExec(t, "creating a temporary table for testing", db,
				"CREATE TABLE migrate_run_test ( name string )")

			sequence := []migration{
				{
					name: "v1",
					run: func(ctx context.NadCtx, db *database.DB) error {
						database.MustExec(t, "marking v1 completed", db, "INSERT INTO migrate_run_test (name) VALUES (?)", "v1")
						return nil
					},
				},
				{
					name: "v2",
					run: func(ctx context.NadCtx, db *database.DB) error {
						database.MustExec(t, "marking v2 completed", db, "INSERT INTO migrate_run_test (name) VALUES (?)", "v2")
						return nil
					},
				},
				{
					name: "v3",
					run: func(ctx context.NadCtx, db *database.DB) error {
						database.MustExec(t, "marking v3 completed", db, "INSERT INTO migrate_run_test (name) VALUES (?)", "v3")
						return nil
					},
				},
			}

			// execute
			err := Run(ctx, sequence, tc.mode)
			if err != nil {
				t.Fatal(errors.Wrap(err, "failed to run"))
			}

			// test
			var schema int
			database.MustScan(t, "getting schema", db.QueryRow("SELECT value FROM system WHERE key = ?", tc.schemaKey), &schema)
			assert.Equal(t, schema, 3, "schema was not updated")

			var testRunCount int
			database.MustScan(t, "counting test runs", db.QueryRow("SELECT count(*) FROM migrate_run_test"), &testRunCount)
			assert.Equal(t, testRunCount, 3, "test run count mismatch")

			var testRun1, testRun2, testRun3 string
			database.MustScan(t, "finding test run 1", db.QueryRow("SELECT name FROM migrate_run_test WHERE name = ?", "v1"), &testRun1)
			database.MustScan(t, "finding test run 2", db.QueryRow("SELECT name FROM migrate_run_test WHERE name = ?", "v2"), &testRun2)
			database.MustScan(t, "finding test run 2", db.QueryRow("SELECT name FROM migrate_run_test WHERE name = ?", "v3"), &testRun3)
		}()
	}
}

func TestRun_up_to_date(t *testing.T) {
	testCases := []struct {
		mode      int
		schemaKey string
	}{
		{
			mode:      LocalMode,
			schemaKey: consts.SystemSchema,
		},
		{
			mode:      RemoteMode,
			schemaKey: consts.SystemRemoteSchema,
		},
	}

	for _, tc := range testCases {
		func() {
			// set up
			opts := database.TestDBOptions{SkipMigration: true}
			ctx := context.InitTestCtx(t, "../tmp", &opts)
			defer context.TeardownTestCtx(t, ctx)

			db := ctx.DB

			database.MustExec(t, "creating a temporary table for testing", db,
				"CREATE TABLE migrate_run_test ( name string )")

			database.MustExec(t, "inserting a schema", db, "INSERT INTO system (key, value) VALUES (?, ?)", tc.schemaKey, 3)

			sequence := []migration{
				{
					name: "v1",
					run: func(ctx context.NadCtx, db *database.DB) error {
						database.MustExec(t, "marking v1 completed", db, "INSERT INTO migrate_run_test (name) VALUES (?)", "v1")
						return nil
					},
				},
				{
					name: "v2",
					run: func(ctx context.NadCtx, db *database.DB) error {
						database.MustExec(t, "marking v2 completed", db, "INSERT INTO migrate_run_test (name) VALUES (?)", "v2")
						return nil
					},
				},
				{
					name: "v3",
					run: func(ctx context.NadCtx, db *database.DB) error {
						database.MustExec(t, "marking v3 completed", db, "INSERT INTO migrate_run_test (name) VALUES (?)", "v3")
						return nil
					},
				},
			}

			// execute
			err := Run(ctx, sequence, tc.mode)
			if err != nil {
				t.Fatal(errors.Wrap(err, "failed to run"))
			}

			// test
			var schema int
			database.MustScan(t, "getting schema", db.QueryRow("SELECT value FROM system WHERE key = ?", tc.schemaKey), &schema)
			assert.Equal(t, schema, 3, "schema was not updated")

			var testRunCount int
			database.MustScan(t, "counting test runs", db.QueryRow("SELECT count(*) FROM migrate_run_test"), &testRunCount)
			assert.Equal(t, testRunCount, 0, "test run count mismatch")
		}()
	}
}
