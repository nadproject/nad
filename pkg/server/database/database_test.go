/* Copyright (C) 2019 Monomax Software Pty Ltd
 *
 * This file is part of NAD.
 *
 * NAD is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * NAD is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with NAD.  If not, see <https://www.gnu.org/licenses/>.
 */

package database

import (
	"github.com/nadproject/nad/pkg/assert"
	"testing"
)

func TestValidateConfig(t *testing.T) {
	testCases := []struct {
		input    Config
		expected error
	}{
		{
			input: Config{
				Host:     "mockHost",
				Port:     "mockPort",
				Name:     "mockName",
				User:     "mockUser",
				Password: "mockPassword",
			},
			expected: nil,
		},
		{
			input: Config{
				Host: "mockHost",
				Port: "mockPort",
				Name: "mockName",
				User: "mockUser",
			},
			expected: nil,
		},
		{
			input: Config{
				Port:     "mockPort",
				Name:     "mockName",
				User:     "mockUser",
				Password: "mockPassword",
			},
			expected: ErrConfigMissingHost,
		},
		{
			input: Config{
				Host:     "mockHost",
				Name:     "mockName",
				User:     "mockUser",
				Password: "mockPassword",
			},
			expected: ErrConfigMissingPort,
		},
		{
			input: Config{
				Host:     "mockHost",
				Port:     "mockPort",
				User:     "mockUser",
				Password: "mockPassword",
			},
			expected: ErrConfigMissingName,
		},
		{
			input: Config{
				Host:     "mockHost",
				Port:     "mockPort",
				Name:     "mockName",
				Password: "mockPassword",
			},
			expected: ErrConfigMissingUser,
		},
		{
			input:    Config{},
			expected: ErrConfigMissingHost,
		},
	}

	for _, tc := range testCases {
		result := validateConfig(tc.input)

		assert.Equal(t, result, tc.expected, "result mismatch")
	}
}
