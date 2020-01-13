package controllers

import "github.com/nadproject/nad/pkg/server/views"

func NewStatic() *Static {
	return &Static{
		Home: views.NewView("base", "static/home"),
		//Contact: views.NewView("base", "static/contact"),
	}
}

type Static struct {
	Home    *views.View
	Contact *views.View
}
