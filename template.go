package bytego

import (
	"io"
	"io/fs"
	"path/filepath"
	"strings"
	"text/template"
)

type defaultTemplae struct {
	ext       string
	templates *template.Template
}

func (t *defaultTemplae) Render(w io.Writer, name string, data interface{}) error {
	if !strings.HasSuffix(name, t.ext) {
		name = name + t.ext
	}
	return t.templates.ExecuteTemplate(w, name, data)
}

func NewTemplate(pattner string, fsys ...fs.FS) Renderer {
	var tmpl *template.Template
	if len(fsys) > 0 {
		tmplFS, err := template.ParseFS(fsys[0], pattner)
		if err != nil {
			panic(err)
		}
		tmpl = tmplFS
	} else {
		tmpl = template.Must(template.ParseGlob(pattner))
	}
	return &defaultTemplae{
		ext:       filepath.Ext(pattner),
		templates: tmpl,
	}
}
