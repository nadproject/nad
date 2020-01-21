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

package app

import (
	"time"

	"github.com/nadproject/nad/pkg/server/crypt"
	"github.com/nadproject/nad/pkg/server/database"
	"github.com/pkg/errors"
)

// CreateSession returns a new session for the user of the given id
func (a *App) CreateSession(userID int) (database.Session, error) {
	key, err := crypt.GetRandomStr(32)
	if err != nil {
		return database.Session{}, errors.Wrap(err, "generating key")
	}

	session := database.Session{
		UserID:     userID,
		Key:        key,
		LastUsedAt: time.Now(),
		ExpiresAt:  time.Now().Add(24 * 100 * time.Hour),
	}

	if err := a.DB.Save(&session).Error; err != nil {
		return database.Session{}, errors.Wrap(err, "saving session")
	}

	return session, nil
}

// DeleteUserSessions deletes all existing sessions for the given user. It effectively
// invalidates all existing sessions.
func (a *App) DeleteUserSessions(userID int) error {
	if err := a.DB.Where("user_id = ?", userID).Delete(&database.Session{}).Error; err != nil {
		return errors.Wrap(err, "deleting sessions")
	}

	return nil
}

// DeleteSession deletes the session that match the given info
func (a *App) DeleteSession(sessionKey string) error {
	if err := a.DB.Where("key = ?", sessionKey).Delete(&database.Session{}).Error; err != nil {
		return errors.Wrap(err, "deleting the session")
	}

	return nil
}
