package theme

import "github.com/fatih/color"

// Style represents a named color style
type Style struct {
	fg      color.Attribute
	bg      color.Attribute
	attrs   []color.Attribute
	printer *color.Color
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
	}
}

// Print prints text using the style
func (s *Style) Print(a ...interface{}) {
	s.printer.Print(a...)
}

// Printf prints formatted text using the style
func (s *Style) Printf(format string, a ...interface{}) {
	s.printer.Printf(format, a...)
}

// Println prints text using the style followed by a newline
func (s *Style) Println(a ...interface{}) {
	s.printer.Println(a...)
}

// Sprint returns styled text as string
func (s *Style) Sprint(a ...interface{}) string {
	return s.printer.Sprint(a...)
}

// Sprintf returns styled formatted text as string
func (s *Style) Sprintf(format string, a ...interface{}) string {
	return s.printer.Sprintf(format, a...)
}

// Sprintln returns styled text with newline as string
func (s *Style) Sprintln(a ...interface{}) string {
	return s.printer.Sprintln(a...)
}
