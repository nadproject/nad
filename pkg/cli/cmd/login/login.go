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

package login

import (
	"strconv"

	"github.com/nadproject/nad/pkg/cli/client"
	"github.com/nadproject/nad/pkg/cli/consts"
	"github.com/nadproject/nad/pkg/cli/context"
	"github.com/nadproject/nad/pkg/cli/database"
	"github.com/nadproject/nad/pkg/cli/infra"
	"github.com/nadproject/nad/pkg/cli/log"
	"github.com/nadproject/nad/pkg/cli/ui"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var example = `
  nad login`

// NewCmd returns a new login command
func NewCmd(ctx context.NadCtx) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "login",
		Short:   "Login to nad server",
		Example: example,
		RunE:    newRun(ctx),
	}

	return cmd
}

// Do dervies credentials on the client side and requests a session token from the server
func Do(ctx context.NadCtx, email, password string) error {
	signinResp, err := client.Signin(ctx, email, password)
	if err != nil {
		return errors.Wrap(err, "requesting session")
	}

	db := ctx.DB
	tx, err := db.Begin()
	if err != nil {
		return errors.Wrap(err, "beginning a transaction")
	}

	if err := database.UpsertSystem(tx, consts.SystemSessionKey, signinResp.Key); err != nil {
		return errors.Wrap(err, "saving session key")
	}
	if err := database.UpsertSystem(tx, consts.SystemSessionKeyExpiry, strconv.FormatInt(signinResp.ExpiresAt, 10)); err != nil {
		return errors.Wrap(err, "saving session key")
	}

	tx.Commit()

	return nil
}

func newRun(ctx context.NadCtx) infra.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		log.Plain("Welcome to NAD Pro (https://www.getnad.com).\n")

		var email, password string
		if err := ui.PromptInput("email", &email); err != nil {
			return errors.Wrap(err, "getting email input")
		}
		if email == "" {
			return errors.New("Email is empty")
		}

		if err := ui.PromptPassword("password", &password); err != nil {
			return errors.Wrap(err, "getting password input")
		}
		if password == "" {
			return errors.New("Password is empty")
		}

		err := Do(ctx, email, password)
		if errors.Cause(err) == client.ErrInvalidLogin {
			log.Error("wrong login\n")
			return nil
		} else if err != nil {
			return errors.Wrap(err, "logging in")
		}

		log.Success("logged in\n")

		return nil
	}

}
