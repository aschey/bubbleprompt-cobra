package cprompt

import (
	"context"

	"github.com/spf13/cobra"
)

const placeholderKey string = "placeholder"

func cmdContext(cmd *cobra.Command) context.Context {
	ctx := cmd.Context()
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

func keyPrefix(cmd *cobra.Command) string {
	if cmd.HasParent() {
		return cmd.Parent().Name() + cmd.Name()
	}
	return cmd.Name()
}

func contextVal(cmd *cobra.Command, key string) any {
	ctx := cmdContext(cmd)
	return ctx.Value(keyPrefix(cmd) + key)
}

func updateContext(cmd *cobra.Command, key string, val any) {
	ctx := cmdContext(cmd)
	ctx = context.WithValue(ctx, keyPrefix(cmd)+key, val)
	cmd.SetContext(ctx)
}

func placeholders(cmd *cobra.Command) []string {
	val := contextVal(cmd, placeholderKey)
	if val == nil {
		return []string{}
	}
	return val.([]string)
}

func SetPlaceholders(cmd *cobra.Command, placeholders ...string) {
	updateContext(cmd, placeholderKey, placeholders)
}
