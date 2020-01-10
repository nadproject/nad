package tpl

import (
	"html/template"
	"log"
	"os"
	"path/filepath"
)

// Templates holds parsed HTML template instances
type Templates map[string]*template.Template

func getWebTemplatePaths(dirpath string) ([]string, error) {
	var paths []string
	err := filepath.Walk(dirpath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	log.Println(paths)

	return paths, nil
}

// ParseWeb parses HTML templates for the web application.
func ParseWeb(templateDir string) Templates {

	resolve := func(filename string) string {
		return filepath.Join(templateDir, filename)
	}

	base := resolve("base.html")

	t := make(map[string]*template.Template)
	t["index.html"] = template.Must(template.ParseFiles(resolve("index.html"), base))
	t["user/join.html"] = template.Must(template.ParseFiles(resolve("user/join.html"), base))

	return t
}
