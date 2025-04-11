package types

import "context"

// CommandFunc defines the function signature for command handlers.
type CommandFunc func(ctx context.Context, args []string) (response string, err error)
