package theme

import (
	"fmt"
	"github.com/fatih/color"
	"io"
	"os"
)

// StylePrinter defines an interface for printing styled text
type StylePrinter interface {
	Print(a ...interface{})
	Printf(format string, a ...interface{})
	Println(a ...interface{})
}

// Style represents a named color style
type Style struct {
	fg      color.Attribute
	bg      color.Attribute
	attrs   []color.Attribute
	printer *color.Color
	writer  io.Writer
}

// NewStyle creates a new style with foreground, background and attributes
func NewStyle(fg, bg color.Attribute, attrs ...color.Attribute) *Style {
	c := color.New(fg)

	if bg != 0 {
		c.Add(bg)
	}

	if len(attrs) > 0 {
		c.Add(attrs...)
	}

	return &Style{
		fg:      fg,
		bg:      bg,
		attrs:   attrs,
		printer: c,
		writer:  os.Stdout,
	}
}

// WithWriter sets a custom writer for the style
func (s *Style) WithWriter(w io.Writer) *Style {
	s.writer = w
	return s
}

// Print prints text using the style
func (s *Style) Print(a ...interface{}) {
	if s.writer == os.Stdout {
		s.printer.Print(a...)
	} else {
		fmt.Fprint(s.writer, s.sprint(a...))
	}
}

// Printf prints formatted text using the style
func (s *Style) Printf(format string, a ...interface{}) {
	if s.writer == os.Stdout {
		s.printer.Printf(format, a...)
	} else {
		fmt.Fprint(s.writer, s.sprintf(format, a...))
	}
}

// Println prints text using the style followed by a newline
func (s *Style) Println(a ...interface{}) {
	if s.writer == os.Stdout {
		s.printer.Println(a...)
	} else {
		fmt.Fprintln(s.writer, s.sprint(a...))
	}
}

// sprint returns styled text as string
func (s *Style) sprint(a ...interface{}) string {
	return s.printer.Sprint(a...)
}

// Sprintf returns styled formatted text as string
func (s *Style) sprintf(format string, a ...interface{}) string {
	return s.printer.Sprintf(format, a...)
}
