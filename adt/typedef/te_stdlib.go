//go:build !goverter

package typedef

import (
	"embed"
	"html/template"
	"log/slog"

	"github.com/Masterminds/sprig/v3"

	"orglang/go-runtime/lib/te"
)

//go:embed all:vp
var teFs embed.FS

func newRendererStdlib(l *slog.Logger) (*te.RendererStdlib, error) {
	t, err := template.New("typedef").Funcs(sprig.FuncMap()).ParseFS(teFs, "vp/bs5/*.html")
	if err != nil {
		return nil, err
	}
	return te.NewRendererStdlib(t, l), nil
}
