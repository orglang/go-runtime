package msg

import (
	"bytes"
	"html/template"
	"log/slog"
)

const (
	MIMEAnyAny  = "*/*"
	MIMETextAny = "text/*"
)

// port
type Renderer interface {
	Render(string, any) ([]byte, error)
}

// adapter
type RendererStdlib struct {
	registry *template.Template
	log      *slog.Logger
}

func NewRendererStdlib(t *template.Template, l *slog.Logger) *RendererStdlib {
	name := slog.String("name", "msg.RendererStdlib")
	return &RendererStdlib{t, l.With(name)}
}

func (r *RendererStdlib) Render(name string, data any) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := r.registry.ExecuteTemplate(buf, name, data)
	if err != nil {
		r.log.Error("rendering failed", slog.Any("reason", err))
	}
	return buf.Bytes(), err
}
