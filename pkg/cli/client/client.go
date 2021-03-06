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

// Package client provides interfaces for interacting with the NAD server
// and the data structures for responses
package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/nadproject/nad/pkg/cli/context"
	"github.com/nadproject/nad/pkg/cli/log"
	"github.com/pkg/errors"
)

// ErrInvalidLogin is an error for invalid credentials for login
var ErrInvalidLogin = errors.New("wrong credentials")

// requestOptions contians options for requests
type requestOptions struct {
	HTTPClient *http.Client
}

func getReq(ctx context.NadCtx, path, method, body string) (*http.Request, error) {
	endpoint := fmt.Sprintf("%s%s", ctx.APIEndpoint, path)
	req, err := http.NewRequest(method, endpoint, strings.NewReader(body))
	if err != nil {
		return nil, errors.Wrap(err, "constructing http request")
	}

	req.Header.Set("CLI-Version", ctx.Version)
	req.Header.Set("Content-Type", "application/json")

	if ctx.SessionKey != "" {
		credential := fmt.Sprintf("Bearer %s", ctx.SessionKey)
		req.Header.Set("Authorization", credential)
	}

	return req, nil
}

// checkRespErr checks if the given http response indicates an error. It returns a boolean indicating
// if the response is an error, and a decoded error message.
func checkRespErr(res *http.Response) error {
	if res.StatusCode < 400 {
		return nil
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return errors.Wrapf(err, "server responded with %d but client could not read the response body", res.StatusCode)
	}

	bodyStr := string(body)
	return errors.Errorf(`response %d "%s"`, res.StatusCode, strings.TrimRight(bodyStr, "\n"))
}

// doReq does a http request to the given path in the api endpoint
func doReq(ctx context.NadCtx, method, path, body string, options *requestOptions) (*http.Response, error) {
	req, err := getReq(ctx, path, method, body)
	if err != nil {
		return nil, errors.Wrap(err, "getting request")
	}

	log.Debug("HTTP request: %+v\n", req)

	var hc http.Client
	if options != nil && options.HTTPClient != nil {
		hc = *options.HTTPClient
	} else {
		hc = http.Client{}
	}

	res, err := hc.Do(req)
	if err != nil {
		return res, errors.Wrap(err, "making http request")
	}

	if err = checkRespErr(res); err != nil {
		return res, errors.Wrap(err, "server responded with an error")
	}

	return res, nil
}

// doAuthorizedReq does a http request to the given path in the api endpoint as a user,
// with the appropriate headers. The given path should include the preceding slash.
func doAuthorizedReq(ctx context.NadCtx, method, path, body string, options *requestOptions) (*http.Response, error) {
	if ctx.SessionKey == "" {
		return nil, errors.New("no session key found")
	}

	return doReq(ctx, method, path, body, options)
}

// GetSyncStateResp is the response get sync state endpoint
type GetSyncStateResp struct {
	FullSyncBefore int   `json:"full_sync_before"`
	MaxUSN         int   `json:"max_usn"`
	CurrentTime    int64 `json:"current_time"`
}

// GetSyncState gets the sync state response from the server
func GetSyncState(ctx context.NadCtx) (GetSyncStateResp, error) {
	var ret GetSyncStateResp

	res, err := doAuthorizedReq(ctx, "GET", "/v1/sync/state", "", nil)
	if err != nil {
		return ret, errors.Wrap(err, "constructing http request")
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return ret, errors.Wrap(err, "reading the response body")
	}

	if err = json.Unmarshal(body, &ret); err != nil {
		return ret, errors.Wrap(err, "unmarshalling the payload")
	}

	return ret, nil
}

// SyncFragNote represents a note in a sync fragment and contains only the necessary information
// for the client to sync the note locally
type SyncFragNote struct {
	UUID      string    `json:"uuid"`
	BookUUID  string    `json:"book_uuid"`
	USN       int       `json:"usn"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	AddedOn   int64     `json:"added_on"`
	EditedOn  int64     `json:"edited_on"`
	Body      string    `json:"content"`
	Public    bool      `json:"public"`
	Deleted   bool      `json:"deleted"`
}

// SyncFragBook represents a book in a sync fragment and contains only the necessary information
// for the client to sync the note locally
type SyncFragBook struct {
	UUID      string    `json:"uuid"`
	USN       int       `json:"usn"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	AddedOn   int64     `json:"added_on"`
	Name      string    `json:"name"`
	Deleted   bool      `json:"deleted"`
}

// SyncFragment contains a piece of information about the server's state.
type SyncFragment struct {
	FragMaxUSN    int            `json:"frag_max_usn"`
	UserMaxUSN    int            `json:"user_max_usn"`
	CurrentTime   int64          `json:"current_time"`
	Notes         []SyncFragNote `json:"notes"`
	Books         []SyncFragBook `json:"books"`
	ExpungedNotes []string       `json:"expunged_notes"`
	ExpungedBooks []string       `json:"expunged_books"`
}

// GetSyncFragmentResp is the response from the get sync fragment endpoint
type GetSyncFragmentResp struct {
	Fragment SyncFragment `json:"fragment"`
}

// GetSyncFragment gets a sync fragment response from the server
func GetSyncFragment(ctx context.NadCtx, afterUSN int) (GetSyncFragmentResp, error) {
	v := url.Values{}
	v.Set("after_usn", strconv.Itoa(afterUSN))
	queryStr := v.Encode()

	path := fmt.Sprintf("/v1/sync/fragment?%s", queryStr)
	res, err := doAuthorizedReq(ctx, "GET", path, "", nil)

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return GetSyncFragmentResp{}, errors.Wrap(err, "reading the response body")
	}

	var resp GetSyncFragmentResp
	if err = json.Unmarshal(body, &resp); err != nil {
		return resp, errors.Wrap(err, "unmarshalling the payload")
	}

	return resp, nil
}

// RespBook is the book in the response from the create book api
type RespBook struct {
	ID        int       `json:"id"`
	UUID      string    `json:"uuid"`
	USN       int       `json:"usn"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name"`
}

// CreateBookPayload is a payload for creating a book
type CreateBookPayload struct {
	Name string `json:"name"`
}

// CreateBook creates a new book in the server
func CreateBook(ctx context.NadCtx, name string) (RespBook, error) {
	payload := CreateBookPayload{
		Name: name,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return RespBook{}, errors.Wrap(err, "marshaling payload")
	}

	res, err := doAuthorizedReq(ctx, "POST", "/v1/books", string(b), nil)
	if err != nil {
		return RespBook{}, errors.Wrap(err, "posting a book to the server")
	}

	var resp RespBook
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return resp, errors.Wrap(err, "decoding response payload")
	}

	return resp, nil
}

type updateBookPayload struct {
	Name *string `json:"name"`
}

// UpdateBookResp is the response from create book api
type UpdateBookResp struct {
	Book RespBook `json:"book"`
}

// UpdateBook updates a book in the server
func UpdateBook(ctx context.NadCtx, name, uuid string) (UpdateBookResp, error) {
	payload := updateBookPayload{
		Name: &name,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return UpdateBookResp{}, errors.Wrap(err, "marshaling payload")
	}

	endpoint := fmt.Sprintf("/v1/books/%s", uuid)
	res, err := doAuthorizedReq(ctx, "PATCH", endpoint, string(b), nil)
	if err != nil {
		return UpdateBookResp{}, errors.Wrap(err, "posting a book to the server")
	}

	var resp UpdateBookResp
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return resp, errors.Wrap(err, "decoding payload")
	}

	return resp, nil
}

// DeleteBookResp is the response from create book api
type DeleteBookResp struct {
	Status int      `json:"status"`
	Book   RespBook `json:"book"`
}

// DeleteBook deletes a book in the server
func DeleteBook(ctx context.NadCtx, uuid string) (DeleteBookResp, error) {
	endpoint := fmt.Sprintf("/v1/books/%s", uuid)
	res, err := doAuthorizedReq(ctx, "DELETE", endpoint, "", nil)
	if err != nil {
		return DeleteBookResp{}, errors.Wrap(err, "deleting a book in the server")
	}

	var resp DeleteBookResp
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return resp, errors.Wrap(err, "decoding the response")
	}

	return resp, nil
}

// CreateNotePayload is a payload for creating a note
type CreateNotePayload struct {
	BookUUID string `json:"book_uuid"`
	Body     string `json:"content"`
}

// CreateNoteResp is the response from create note endpoint
type CreateNoteResp struct {
	Result RespNote `json:"result"`
}

type respNoteBook struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
}

type respNoteUser struct {
	Name string `json:"name"`
}

// RespNote is a note in the response
type RespNote struct {
	UUID      string       `json:"uuid"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
	Body      string       `json:"content"`
	AddedOn   int64        `json:"added_on"`
	Public    bool         `json:"public"`
	USN       int          `json:"usn"`
	Book      respNoteBook `json:"book"`
	User      respNoteUser `json:"user"`
}

// CreateNote creates a note in the server
func CreateNote(ctx context.NadCtx, bookUUID, content string) (CreateNoteResp, error) {
	payload := CreateNotePayload{
		BookUUID: bookUUID,
		Body:     content,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return CreateNoteResp{}, errors.Wrap(err, "marshaling payload")
	}

	res, err := doAuthorizedReq(ctx, "POST", "/v1/notes", string(b), nil)
	if err != nil {
		return CreateNoteResp{}, errors.Wrap(err, "posting a book to the server")
	}

	var resp CreateNoteResp
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return CreateNoteResp{}, errors.Wrap(err, "decoding payload")
	}

	return resp, nil
}

type updateNotePayload struct {
	BookUUID *string `json:"book_uuid"`
	Body     *string `json:"content"`
	Public   *bool   `json:"public"`
}

// UpdateNoteResp is the response from create book api
type UpdateNoteResp struct {
	Status int      `json:"status"`
	Result RespNote `json:"result"`
}

// UpdateNote updates a note in the server
func UpdateNote(ctx context.NadCtx, uuid, bookUUID, content string, public bool) (UpdateNoteResp, error) {
	payload := updateNotePayload{
		BookUUID: &bookUUID,
		Body:     &content,
		Public:   &public,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return UpdateNoteResp{}, errors.Wrap(err, "marshaling payload")
	}

	endpoint := fmt.Sprintf("/v1/notes/%s", uuid)
	res, err := doAuthorizedReq(ctx, "PATCH", endpoint, string(b), nil)
	if err != nil {
		return UpdateNoteResp{}, errors.Wrap(err, "patching a note to the server")
	}

	var resp UpdateNoteResp
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return UpdateNoteResp{}, errors.Wrap(err, "decoding payload")
	}

	return resp, nil
}

// DeleteNoteResp is the response from remove note api
type DeleteNoteResp struct {
	Status int      `json:"status"`
	Result RespNote `json:"result"`
}

// DeleteNote removes a note in the server
func DeleteNote(ctx context.NadCtx, uuid string) (DeleteNoteResp, error) {
	endpoint := fmt.Sprintf("/v1/notes/%s", uuid)
	res, err := doAuthorizedReq(ctx, "DELETE", endpoint, "", nil)
	if err != nil {
		return DeleteNoteResp{}, errors.Wrap(err, "patching a note to the server")
	}

	var resp DeleteNoteResp
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return DeleteNoteResp{}, errors.Wrap(err, "decoding payload")
	}

	return resp, nil
}

// GetBooksResp is a response from get books endpoint
type GetBooksResp []struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
}

// GetBooks gets books from the server
func GetBooks(ctx context.NadCtx, sessionKey string) (GetBooksResp, error) {
	res, err := doAuthorizedReq(ctx, "GET", "/v1/books", "", nil)
	if err != nil {
		return GetBooksResp{}, errors.Wrap(err, "making http request")
	}

	var resp GetBooksResp
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return GetBooksResp{}, errors.Wrap(err, "decoding payload")
	}

	return resp, nil
}

// SigninPayload is a payload for /v1/login
type SigninPayload struct {
	Email    string `json:"email"`
	Passowrd string `json:"password"`
}

// SigninResponse is a response from /v1/login endpoint
type SigninResponse struct {
	Key       string `json:"key"`
	ExpiresAt int64  `json:"expires_at"`
}

// Signin requests a session token
func Signin(ctx context.NadCtx, email, password string) (SigninResponse, error) {
	payload := SigninPayload{
		Email:    email,
		Passowrd: password,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return SigninResponse{}, errors.Wrap(err, "marshaling payload")
	}
	res, err := doReq(ctx, "POST", "/v1/login", string(b), nil)
	if err != nil {
		if res.StatusCode == http.StatusBadRequest {
			return SigninResponse{}, ErrInvalidLogin
		}

		return SigninResponse{}, errors.Wrap(err, "making http request")
	}

	var resp SigninResponse
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return SigninResponse{}, errors.Wrap(err, "decoding payload")
	}

	return resp, nil
}

// Signout deletes a user session on the server side
func Signout(ctx context.NadCtx, sessionKey string) error {
	hc := http.Client{
		// No need to follow redirect
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	opts := requestOptions{
		HTTPClient: &hc,
	}
	_, err := doAuthorizedReq(ctx, "POST", "/v1/logout", "", &opts)
	if err != nil {
		return errors.Wrap(err, "making http request")
	}

	return nil
}
