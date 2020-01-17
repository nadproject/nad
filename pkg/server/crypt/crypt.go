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

package crypt

import (
	"crypto/rand"
	"crypto/sha256"

	"encoding/base64"
	"github.com/pkg/errors"
	"golang.org/x/crypto/pbkdf2"
)

// ServerKDFIteration is the iteration count for PBKDF on the server
var ServerKDFIteration = 100000

// RandomBytes generates a cryptographically secure pseudorandom bytes of the
// given size
func RandomBytes(numBytes int) ([]byte, error) {
	b := make([]byte, numBytes)
	if _, err := rand.Read(b); err != nil {
		return nil, errors.Wrap(err, "reading random bits")
	}

	return b, nil
}

// GetRandomStr generates a cryptographically secure pseudorandom bytes of the
// given size, encoded in base64
func GetRandomStr(numBytes int) (string, error) {
	b, err := RandomBytes(numBytes)
	if err != nil {
		return "", errors.Wrap(err, "generating random bits")
	}

	return base64.StdEncoding.EncodeToString(b), nil
}

// HashAuthKey hashes the authKey provided by a client
func HashAuthKey(authKey, salt string, iteration int) string {
	keyHashBits := pbkdf2.Key([]byte(authKey), []byte(salt), iteration, 32, sha256.New)

	return base64.StdEncoding.EncodeToString(keyHashBits)
}
