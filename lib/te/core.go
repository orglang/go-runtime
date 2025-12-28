package te

const (
	MIMEAnyAny  = "*/*"
	MIMETextAny = "text/*"
)

type Renderer interface {
	Render(string, any) ([]byte, error)
}
