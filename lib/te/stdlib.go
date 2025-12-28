package te

import (
	"bytes"
	"html/template"
	"log/slog"
)

type RendererStdlib struct {
	te  *template.Template
	log *slog.Logger
}

func NewRendererStdlib(t *template.Template, l *slog.Logger) *RendererStdlib {
	name := slog.String("name", "te.RendererStdlib")
	return &RendererStdlib{t, l.With(name)}
}

func (r *RendererStdlib) Render(name string, data any) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := r.te.ExecuteTemplate(buf, name, data)
	if err != nil {
		r.log.Error("rendering failed", slog.Any("reason", err))
	}
	return buf.Bytes(), err
}
