package theme

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
