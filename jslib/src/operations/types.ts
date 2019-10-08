/* Copyright (C) 2019 Monomax Software Pty Ltd
 *
 * This file is part of Dnote.
 *
 * Dnote is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * Dnote is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with Dnote.  If not, see <https://www.gnu.org/licenses/>.
 */

// NoteData represents a data for a note as returned by services.
// The response from services need to conform to this interface.
export interface NoteData {
  uuid: string;
  created_at: string;
  updated_at: string;
  content: string;
  added_on: number;
  public: boolean;
  usn: number;
  book: {
    uuid: string;
    label: string;
  };
  user: {
    name: string;
    uuid: string;
  };
}

export interface EmailPrefData {
  digestWeekly: boolean;
}

export interface UserData {
  uuid: string;
  email: string;
  emailVerified: boolean;
  pro: boolean;
  classic: boolean;
}

export type BookData = {
  uuid: string;
  usn: number;
  created_at: string;
  updated_at: string;
  label: string;
};

export interface DigestRuleData {
  uuid: string;
  title: string;
  enabled: boolean;
  hour: number;
  minute: number;
  frequency: number;
  books: BookData[];
  created_at: number;
  updated_at: number;
}
