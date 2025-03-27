package theme

import (
	"fmt"
	"io"
)

// Writer defines an interface for writing output
type Writer interface {
	Print(a ...interface{})
	Printf(format string, a ...interface{})
	Println(a ...interface{})
}

// StdoutWriter implements the Writer interface for standard output
type StdoutWriter struct{}

// Print writes to stdout
func (w *StdoutWriter) Print(a ...interface{}) {
	fmt.Print(a...)
}

// Printf writes formatted output to stdout
func (w *StdoutWriter) Printf(format string, a ...interface{}) {
	fmt.Printf(format, a...)
}

// Println writes to stdout with a newline
func (w *StdoutWriter) Println(a ...interface{}) {
	fmt.Println(a...)
}

// IOWriter adapts an io.Writer to the Writer interface
type IOWriter struct {
	Writer io.Writer
}

// Print writes to the underlying io.Writer
func (w *IOWriter) Print(a ...interface{}) {
	fmt.Fprint(w.Writer, a...)
}

// Printf writes formatted output to the underlying io.Writer
func (w *IOWriter) Printf(format string, a ...interface{}) {
	fmt.Fprintf(w.Writer, format, a...)
}

// Println writes to the underlying io.Writer with a newline
func (w *IOWriter) Println(a ...interface{}) {
	fmt.Fprintln(w.Writer, a...)
}
