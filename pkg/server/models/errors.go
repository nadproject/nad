package models

import (
	"strings"
)

type modelError string

var (
	// ErrNotFound an error that indicates that the given resource is not found
	ErrNotFound notFoundError = notFoundError{"not found"}
	// ErrEmailDuplicate is an error for duplicate email
	ErrEmailDuplicate conflictError = conflictError{"duplicate email exists"}
	// ErrLoginInvalid is an error for invalid login
	ErrLoginInvalid badRequestError = badRequestError{"invalid login"}
	// ErrPasswordTooShort is an error for password being too short
	ErrPasswordTooShort badRequestError = badRequestError{"password too short"}
	// ErrEmailRequired is an error for missing email
	ErrEmailRequired badRequestError = badRequestError{"email is required"}
	// ErrEmailInvalid is an error for invalid email
	ErrEmailInvalid badRequestError = badRequestError{"email is invalid"}
	// ErrPasswordRequired is an error for missing password
	ErrPasswordRequired badRequestError = badRequestError{"password is required"}
	// ErrSessionKeyRequired is an error for missing session key
	ErrSessionKeyRequired badRequestError = badRequestError{"session key is required"}
	// ErrIDInvalid is an error for missing session key
	ErrIDInvalid badRequestError = badRequestError{"invalid id"}
	// ErrSessionUserIDRequired is an error for missing session key
	ErrSessionUserIDRequired badRequestError = badRequestError{"user_id is required"}

	// ErrNoteUUIDRequired is an error for missing session key
	ErrNoteUUIDRequired badRequestError = badRequestError{"note uuid is required"}
	// ErrNoteUserIDRequired is an error for missing user_id in note
	ErrNoteUserIDRequired badRequestError = badRequestError{"note user_id is required"}
	// ErrNoteBookUUIDRequired is an error for missing book_uuid in note
	ErrNoteBookUUIDRequired badRequestError = badRequestError{"note book_uuid is required"}
	// ErrNoteAddedOnRequired is an error for missing added_on in note
	ErrNoteAddedOnRequired badRequestError = badRequestError{"note added_on is required"}
	// ErrNoteEditedOnRequired is an error for missing edited_on in note
	ErrNoteEditedOnRequired badRequestError = badRequestError{"note edited_on is required"}
	// ErrNoteUSNRequired is an error for missing usn in note
	ErrNoteUSNRequired badRequestError = badRequestError{"note usn is required"}

	// ErrBookUUIDRequired is an error for missing session key
	ErrBookUUIDRequired badRequestError = badRequestError{"book uuid is required"}
	// ErrBookUserIDRequired is an error for missing user_id in book
	ErrBookUserIDRequired badRequestError = badRequestError{"book user_id is required"}
	// ErrBookEditedOnRequired is an error for missing edited_on in book
	ErrBookEditedOnRequired badRequestError = badRequestError{"book edited_on is required"}
	// ErrBookUSNRequired is an error for missing usn in book
	ErrBookUSNRequired badRequestError = badRequestError{"book usn is required"}
	// ErrBookNameTaken is an error for book name taken
	ErrBookNameTaken conflictError = conflictError{"book name is taken"}
)

// Error returns a string repsentation of the error.
func (e modelError) Error() string {
	return string(e)
}

// Public returns a string repsentation of the error suitable
// for returning to the client.
func (e modelError) Public() string {
	s := string(e)
	return strings.Title(s)
}

type badRequestError struct {
	modelError
}

// IsBadRequest implements that the error is a bad request
func (e badRequestError) IsBadRequest() bool {
	return true
}

type conflictError struct {
	modelError
}

// IsConflictError indicates that the error should return http status code conflict
func (e conflictError) IsConflictError() bool {
	return true
}

type notFoundError struct {
	modelError
}

// IsNotFoundError indicates that the error should return http status code conflict
func (e notFoundError) IsNotFoundError() bool {
	return true
}

type privateError string

// Error returns a string repsentation of the error.
func (e privateError) Error() string {
	return string(e)
}
