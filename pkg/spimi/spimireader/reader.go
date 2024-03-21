package spimireader

type Chunk interface {
	Read() ([]Document, error)
}

type Document struct {
	No      string
	Content string
}
