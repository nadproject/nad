package controllers

import "github.com/nadproject/nad/pkg/server/views"

// NewStatic creates a new Static controller.
func NewStatic() *Static {
	return &Static{
		// Home: views.NewView("Home", "base", "static/home"),
	}
}

// Static is a static controller
type Static struct {
	Home *views.View
}
