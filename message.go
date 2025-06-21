package gqfsl

type Message struct {
	From     string
	To       string
	Subject  string
	TypeBody string // "text/plain" or "text/html", default "text/plain"
	Body     string
}
