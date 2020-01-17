package models

import "strings"

type modelError string

const (
	// ErrNotFound an error that indicates that the given resource is not found
	ErrNotFound modelError = "not found"
	// ErrLoginInvalid is an error for invalid login
	ErrLoginInvalid modelError = "invalid login"
	// ErrPasswordTooShort is an error for password being too short
	ErrPasswordTooShort modelError = "password too short"
	// ErrEmailRequired is an error for missing email
	ErrEmailRequired modelError = "email is required"
	// ErrEmailInvalid is an error for invalid email
	ErrEmailInvalid modelError = "email is invalid"
	// ErrEmailDuplicate is an error for duplicate email
	ErrEmailDuplicate modelError = "duplicate email exists"
	// ErrPasswordRequired is an error for missing password
	ErrPasswordRequired modelError = "password is required"
	// ErrSessionKeyRequired is an error for missing session key
	ErrSessionKeyRequired modelError = "session key is required"
	// ErrIDInvalid is an error for missing session key
	ErrIDInvalid modelError = "invalid id"
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

type privateError string

// Error returns a string repsentation of the error.
func (e privateError) Error() string {
	return string(e)
}
