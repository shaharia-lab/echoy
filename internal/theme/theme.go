package theme

import (
	"github.com/fatih/color"
)

// Style represents a named color style
type Style struct {
	fg      color.Attribute
	bg      color.Attribute
	attrs   []color.Attribute
	printer *color.Color
}

// Theme defines the interface for theming in the application
type Theme interface {
	// Primary returns the primary style
	Primary() *Style

	// Secondary returns the secondary style
	Secondary() *Style

	// Success returns the success style
	Success() *Style

	// Error returns the error style
	Error() *Style

	// Warning returns the warning style
	Warning() *Style

	// Info returns the info style
	Info() *Style

	// Subtle returns the subtle style
	Subtle() *Style

	// Disabled returns the disabled style
	Disabled() *Style

	// Custom returns a custom style by name
	Custom(name string) *Style

	// IsEnabled reports if colors are enabled
	IsEnabled() bool

	// SetEnabled enables or disables color output
	SetEnabled(enabled bool)
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
