package views

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gorilla/csrf"
	"github.com/nadproject/nad/pkg/server/build"
	"github.com/nadproject/nad/pkg/server/context"
	"github.com/nadproject/nad/pkg/server/log"
	"github.com/pkg/errors"
)

const (
	// templateExt is the template extension
	templateExt string = ".gohtml"
)

const (
	siteTitle = "NAD"
)

// Config is a view config
type Config struct {
	Title          string
	Layout         string
	HeaderTemplate string
}

func (c Config) getLayout() string {
	if c.Layout == "" {
		return "base"
	}

	return c.Layout
}

// NewView returns a new view by parsing  the given layout and files
func NewView(baseDir string, c Config, files ...string) *View {
	addTemplatePath(baseDir, files)
	addTemplateExt(files)
	files = append(files, layoutFiles(baseDir)...)

	t, err := template.New(c.Title).Funcs(template.FuncMap{
		"csrfField": func() (template.HTML, error) {
			return "", errors.New("csrfField is not implemented")
		},
		"css": func() []string {
			return strings.Split(build.CSSFiles, ",")
		},
		"title": func() string {
			if c.Title != "" {
				return fmt.Sprintf("%s | %s", c.Title, siteTitle)
			}

			return siteTitle
		},
		"headerTemplate": func() string {
			return c.HeaderTemplate
		},
	}).ParseFiles(files...)
	if err != nil {
		panic(errors.Wrap(err, "instantiating view."))
	}

	return &View{
		Template: t,
		Layout:   c.getLayout(),
	}
}

// View holds the information about a view
type View struct {
	Template *template.Template
	Layout   string
}

func (v *View) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	v.Render(w, r, nil)
}

// Render is used to render the view with the predefined layout.
func (v *View) Render(w http.ResponseWriter, r *http.Request, data interface{}) {
	w.Header().Set("Content-Type", "text/html")

	var vd Data
	switch d := data.(type) {
	case Data:
		vd = d
		// do nothing
	default:
		vd = Data{
			Yield: data,
		}
	}

	if alert := getAlert(r); alert != nil {
		vd.Alert = alert
		clearAlert(w)
	}

	vd.User = context.User(r.Context())

	var buf bytes.Buffer
	csrfField := csrf.TemplateField(r)
	tpl := v.Template.Funcs(template.FuncMap{
		"csrfField": func() template.HTML {
			return csrfField
		},
	})

	if err := tpl.ExecuteTemplate(&buf, v.Layout, vd); err != nil {
		log.ErrorWrap(err, fmt.Sprintf("executing a template '%s'", v.Template.Name()))
		http.Error(w, AlertMsgGeneric, http.StatusInternalServerError)
		return
	}
	io.Copy(w, &buf)
}

// layoutFiles returns a slice of strings representing
// the layout files used in our application.
func layoutFiles(baseDir string) []string {
	pattern := fmt.Sprintf("%s/layouts/*%s", baseDir, templateExt)

	files, err := filepath.Glob(pattern)
	if err != nil {
		panic(err)
	}

	return files
}

// addTemplatePath takes in a slice of strings
// representing file paths for templates.
func addTemplatePath(baseDir string, files []string) {
	for i, f := range files {
		files[i] = fmt.Sprintf("%s/%s", baseDir, f)
	}
}

// addTemplateExt takes in a slice of strings
// representing file paths for templates and it appends
// the templateExt extension to each string in the slice
//
// Eg the input {"home"} would result in the output
// {"home.gohtml"} if templateExt == ".gohtml"
func addTemplateExt(files []string) {
	for i, f := range files {
		files[i] = f + templateExt
	}
}
