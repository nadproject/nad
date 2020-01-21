package controllers

import (
	"os"
	"testing"

	"github.com/nadproject/nad/pkg/server/config"
	"github.com/nadproject/nad/pkg/server/models"
)

const (
	testPageDir = "../views"
)

func TestMain(m *testing.M) {
	cfg := config.Load()

	err := models.InitTestService(cfg)
	if err != nil {
		os.Exit(1)
	}

	code := m.Run()
	os.Exit(code)
}
