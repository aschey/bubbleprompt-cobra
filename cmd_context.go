package cprompt

import (
	"context"

	"github.com/spf13/cobra"
)

var cmdCtx context.Context = context.Background()

const placeholderKey string = "placeholder"

func CmdContext() context.Context {
	return cmdCtx
}

func placeholders(cmd *cobra.Command) []string {
	val := cmdCtx.Value(cmd.Name() + placeholderKey)
	if val == nil {
		return []string{}
	}
	return val.([]string)
}

func SetPlaceholders(cmd *cobra.Command, placeholders ...string) {
	cmdCtx = context.WithValue(cmdCtx, cmd.Name()+placeholderKey, placeholders)
}
