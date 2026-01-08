package templates

import (
	"bytes"
	"html/template"
	"log/slog"
	"path/filepath"
)

// stores pointer to a template set
type Renderer struct {
	templates *template.Template
}

// Looks inside basePath with any file ending with .html
// reads those files and parses them
func NewRenderer(basePath string) (*Renderer, error) {
	tmpl, err := template.ParseGlob(
		filepath.Join(basePath, "*.html"),
	)
	if err != nil {
		slog.Error("Error in html templates", "err", err)
		return nil, err
	}

	return &Renderer{templates: tmpl}, nil
}

// takes name of the html file and data to inject into it

func (r *Renderer) Render(name string, data any) (string, error) {
	var buf bytes.Buffer

	if err := r.templates.ExecuteTemplate(&buf, name, data); err != nil {
		slog.Error("Error filling the template", "err", err)

		return "", err
	}

	return buf.String(), nil
}
