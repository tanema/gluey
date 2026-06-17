package gluey

import (
	"log"

	"github.com/k0kubun/go-ansi"
	"github.com/tanema/gluey/term"
)

// Ctx allows use to keep a root object for all elements
type Ctx struct {
	*log.Logger
	Indent int
}

// New builds a new UI context that every element will be based on
func New() *Ctx {
	return &Ctx{Logger: log.New(ansi.NewAnsiStdout(), "", 0)}
}

// Fmt will format a string template with color and icons
func Fmt(template string, data any) string {
	return term.Sprintf(template, data)
}
