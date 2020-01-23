package controllers

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/nadproject/nad/pkg/assert"
	"github.com/nadproject/nad/pkg/server/config"
	"github.com/nadproject/nad/pkg/server/models"
)

func TestRegister(t *testing.T) {
	testCases := []struct {
		onPremise   bool
		expectedPro bool
	}{
		{
			onPremise:   true,
			expectedPro: true,
		},
		{
			onPremise:   false,
			expectedPro: false,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("self hosting %t", tc.onPremise), func(t *testing.T) {
			cfg := config.Load()
			cfg.OnPremise = tc.onPremise
			cfg.SetPageTemplateDir(testPageDir)
			defer models.ClearTestData(t, models.TestServices.DB)

			usersC := NewUsers(cfg, models.TestServices.User, models.TestServices.Session)

			form := url.Values{}
			form.Add("email", "alice@example.com")
			form.Add("password", "pass1234")
			req := newReq(t, "POST", "/register", form.Encode())
			req.PostForm = form

			w := httpDo(t, usersC.Create, req, nil)

			assert.Equal(t, w.Code, http.StatusFound, "status code mismatch")

			var userCount int
			var userRecord models.User
			models.MustExec(t, models.TestServices.DB.Model(&models.User{}).Count(&userCount), "counting user")
			models.MustExec(t, models.TestServices.DB.First(&userRecord), "finding user")

			assert.Equal(t, userCount, 1, "book count mismatch")
			assert.Equal(t, userRecord.Pro, tc.expectedPro, "user pro mismatch")
		})
	}
}
