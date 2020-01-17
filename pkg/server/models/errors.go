package models

import "strings"

type modelError string

type badRequestError struct {
	modelError
}
type conflictError struct {
	modelError
}

// IsConflictError indicates that the error should return http status code conflict
func (e conflictError) IsConflictError() bool {
	return true
}

const (
	// ErrNotFound an error that indicates that the given resource is not found
	ErrNotFound modelError = "not found"
)

var (
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

// IsBadRequest implements that the error is a bad request
func (e badRequestError) IsBadRequest() bool {
	return true
}

type privateError string

// Error returns a string repsentation of the error.
func (e privateError) Error() string {
	return string(e)
}
