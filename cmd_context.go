package cprompt

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

const modelKey string = "model"

type ctxKey string

func cmdContext(cmd *cobra.Command) context.Context {
	ctx := cmd.Context()
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

func keyPrefix(cmd *cobra.Command) string {
	// Include parent name to ensure key is unique in case of duplicate command names
	if cmd.HasParent() {
		return cmd.Parent().Name() + cmd.Name()
	}
	return cmd.Name()
}

func contextVal(cmd *cobra.Command, key string) any {
	ctx := cmdContext(cmd)
	return ctx.Value(ctxKey(keyPrefix(cmd) + key))
}

func updateContext(cmd *cobra.Command, key string, val any) {
	ctx := cmdContext(cmd)
	ctx = context.WithValue(ctx, ctxKey(keyPrefix(cmd)+key), val)
	cmd.SetContext(ctx)
}

func PreservePlaceholder(cmd *cobra.Command, flag string) {
	updateContext(cmd, flag+"preserve", true)
}

func getPreservePlaceholder(cmd *cobra.Command, flag string) bool {
	val := contextVal(cmd, flag+"preserve")
	if val == nil {
		return false
	}

	return val.(bool)
}

func model(cmd *cobra.Command) tea.Model {
	val := contextVal(cmd, modelKey)
	if val == nil {
		return nil
	}
	return val.(tea.Model)
}

func setModel(cmd *cobra.Command, model tea.Model) {
	updateContext(cmd, modelKey, model)
}
