package cprompt

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

const placeholderKey string = "placeholder"
const modelKey string = "model"
const interactiveKey = "interactive"

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
	return ctx.Value(keyPrefix(cmd) + key)
}

func updateContext(cmd *cobra.Command, key string, val any) {
	ctx := cmdContext(cmd)
	ctx = context.WithValue(ctx, keyPrefix(cmd)+key, val)
	cmd.SetContext(ctx)
}

func setInteractive(cmd *cobra.Command) {
	updateContext(cmd, interactiveKey, true)
}

func interactive(cmd *cobra.Command) bool {
	val := contextVal(cmd, interactiveKey)
	if val == nil {
		return false
	}
	return val.(bool)
}

func placeholders(cmd *cobra.Command) []string {
	val := contextVal(cmd, placeholderKey)
	if val == nil {
		return []string{}
	}
	strVal := val.([]string)
	return strVal
}

func SetPlaceholders(cmd *cobra.Command, placeholders ...string) {
	updateContext(cmd, placeholderKey, placeholders)
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
