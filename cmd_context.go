package cprompt

import (
	"context"

	"github.com/spf13/cobra"
)

const placeholderKey string = "placeholder"

func placeholders(cmd *cobra.Command) []string {
	ctx := cmd.Context()
	if ctx == nil {
		return []string{}
	}
	val := ctx.Value(cmd.Name() + placeholderKey)
	if val == nil {
		return []string{}
	}
	return val.([]string)
}

func SetPlaceholders(cmd *cobra.Command, placeholders ...string) {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	ctx = context.WithValue(ctx, cmd.Name()+placeholderKey, placeholders)
	cmd.SetContext(ctx)
}
