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

// Package consts provides definitions of constants
package consts

var (
	// NADDirName is the name of the directory containing nad files
	NADDirName = ".nad"
	// NADDBFileName is a filename for the NAD SQLite database
	NADDBFileName = "nad.db"
	// TmpContentFileBase is the base for the filename for a temporary content
	TmpContentFileBase = "NAD_TMPCONTENT"
	// TmpContentFileExt is the extension for the temporary content file
	TmpContentFileExt = "md"
	// ConfigFilename is the name of the config file
	ConfigFilename = "nadrc"

	// SystemSchema is the key for schema in the system table
	SystemSchema = "schema"
	// SystemRemoteSchema is the key for remote schema in the system table
	SystemRemoteSchema = "remote_schema"
	// SystemLastSyncAt is the timestamp of the server at the last sync
	SystemLastSyncAt = "last_sync_time"
	// SystemLastMaxUSN is the user's max_usn from the server at the alst sync
	SystemLastMaxUSN = "last_max_usn"
	// SystemLastUpgrade is the timestamp at which the system more recently checked for an upgrade
	SystemLastUpgrade = "last_upgrade"
	// SystemSessionKey is the session key
	SystemSessionKey = "session_token"
	// SystemSessionKeyExpiry is the timestamp at which the session key will expire
	SystemSessionKeyExpiry = "session_token_expiry"
)
