package dec

import (
	"embed"
	"html/template"
	"log/slog"

	"github.com/Masterminds/sprig/v3"

	"orglang/orglang/lib/te"
)

//go:embed all:vp
var vpFs embed.FS

func newRendererStdlib(l *slog.Logger) (*te.RendererStdlib, error) {
	t, err := template.New("proc/dec").Funcs(sprig.FuncMap()).ParseFS(vpFs, "vp/bs5/*.html")
	if err != nil {
		return nil, err
	}
	return te.NewRendererStdlib(t, l), nil
}
