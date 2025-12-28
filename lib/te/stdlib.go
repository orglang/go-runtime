package te

import (
	"bytes"
	"html/template"
	"log/slog"
)

type rendererStdlib struct {
	registry *template.Template
	log      *slog.Logger
}

func newRendererStdlib(t *template.Template, l *slog.Logger) *rendererStdlib {
	name := slog.String("name", "rendererStdlib")
	return &rendererStdlib{t, l.With(name)}
}

func (r *rendererStdlib) Render(name string, data any) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := r.registry.ExecuteTemplate(buf, name, data)
	if err != nil {
		r.log.Error("rendering failed", slog.Any("reason", err))
	}
	return buf.Bytes(), err
}
