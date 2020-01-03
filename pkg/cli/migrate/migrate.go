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
	"database/sql"
	"github.com/nadproject/nad/pkg/cli/consts"
	"github.com/nadproject/nad/pkg/cli/context"
	"github.com/nadproject/nad/pkg/cli/log"
	"github.com/pkg/errors"
)

const (
	// LocalMode is a local migration mode
	LocalMode = iota
	// RemoteMode is a remote migration mode
	RemoteMode
)

// LocalSequence is a list of local migrations to be run
var LocalSequence = []migration{
	lm1,
}

// RemoteSequence is a list of remote migrations to be run
var RemoteSequence = []migration{}

func initSchema(ctx context.NadCtx, schemaKey string) (int, error) {
	// schemaVersion is the index of the latest run migration in the sequence
	schemaVersion := 0

	db := ctx.DB
	_, err := db.Exec("INSERT INTO system (key, value) VALUES (?, ?)", schemaKey, schemaVersion)
	if err != nil {
		return schemaVersion, errors.Wrap(err, "inserting schema")
	}

	return schemaVersion, nil
}

func getSchemaKey(mode int) (string, error) {
	if mode == LocalMode {
		return consts.SystemSchema, nil
	}

	if mode == RemoteMode {
		return consts.SystemRemoteSchema, nil
	}

	return "", errors.Errorf("unsupported migration type '%d'", mode)
}

func getSchema(ctx context.NadCtx, schemaKey string) (int, error) {
	var ret int

	db := ctx.DB
	err := db.QueryRow("SELECT value FROM system where key = ?", schemaKey).Scan(&ret)
	if err == sql.ErrNoRows {
		ret, err = initSchema(ctx, schemaKey)

		if err != nil {
			return ret, errors.Wrap(err, "initializing schema")
		}
	} else if err != nil {
		return ret, errors.Wrap(err, "querying schema")
	}

	return ret, nil
}

func execute(ctx context.NadCtx, m migration, schemaKey string) error {
	log.Debug("running migration %s\n", m.name)

	tx, err := ctx.DB.Begin()
	if err != nil {
		return errors.Wrap(err, "beginning a transaction")
	}

	err = m.run(ctx, tx)
	if err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "running '%s'", m.name)
	}

	var currentSchema int
	err = tx.QueryRow("SELECT value FROM system WHERE key = ?", schemaKey).Scan(&currentSchema)
	if err != nil {
		tx.Rollback()
		return errors.Wrap(err, "getting current schema")
	}

	_, err = tx.Exec("UPDATE system SET value = value + 1 WHERE key = ?", schemaKey)
	if err != nil {
		tx.Rollback()
		return errors.Wrap(err, "incrementing schema")
	}

	tx.Commit()

	return nil
}

// Run performs unrun migrations
func Run(ctx context.NadCtx, migrations []migration, mode int) error {
	schemaKey, err := getSchemaKey(mode)
	if err != nil {
		return errors.Wrap(err, "getting schema key")
	}

	schema, err := getSchema(ctx, schemaKey)
	if err != nil {
		return errors.Wrap(err, "getting the current schema")
	}

	log.Debug("current schema: %s %d of %d\n", consts.SystemSchema, schema, len(migrations))

	toRun := migrations[schema:]

	for _, m := range toRun {
		if err := execute(ctx, m, schemaKey); err != nil {
			return errors.Wrap(err, "running migration")
		}
	}

	return nil
}
