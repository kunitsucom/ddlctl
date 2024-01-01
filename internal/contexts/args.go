package contexts

import (
	"context"
	"os"
)

type contextKeyArgs struct{}

func Args(ctx context.Context) []string {
	if v, ok := ctx.Value(contextKeyArgs{}).([]string); ok {
		return v
	}

	return os.Args[0:]
}

func WithArgs(ctx context.Context, args []string) context.Context {
	return context.WithValue(ctx, contextKeyArgs{}, args)
}
