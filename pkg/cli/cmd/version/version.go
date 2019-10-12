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

package version

import (
	"fmt"

	"github.com/nadproject/nad/pkg/cli/context"
	"github.com/spf13/cobra"
)

// NewCmd returns a new version command
func NewCmd(ctx context.NADCtx) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number of NAD",
		Long:  "Print the version number of NAD",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("nad %s\n", ctx.Version)
		},
	}

	return cmd
}
