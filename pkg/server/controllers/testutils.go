package controllers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nadproject/nad/pkg/server/context"
	"github.com/nadproject/nad/pkg/server/models"
)

func newReq(t *testing.T, method, path, data string) *http.Request {
	return httptest.NewRequest(method, path, strings.NewReader(data))
}

func httpDo(t *testing.T, handler http.HandlerFunc, r *http.Request, user *models.User) *httptest.ResponseRecorder {
	// If a user is provided, set the user in the request context
	if user != nil {
		ctx := r.Context()
		ctx = context.WithUser(ctx, user)
		r = r.WithContext(ctx)
	}

	w := httptest.NewRecorder()
	handler(w, r)

	return w
}
