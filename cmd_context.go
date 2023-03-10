package cprompt

import (
	"context"

	"github.com/aschey/bubbleprompt/input/commandinput"
	"github.com/aschey/bubbleprompt/suggestion"
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

func contextVal(cmd *cobra.Command, key string) any {
	ctx := cmdContext(cmd)
	return ctx.Value(ctxKey(key))
}

func updateContext(cmd *cobra.Command, key string, val any) {
	ctx := cmdContext(cmd)
	ctx = context.WithValue(ctx, ctxKey(key), val)
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

func Completer[T any](cmd *cobra.Command, f func(cmd *cobra.Command, args []string,
	toComplete string) ([]suggestion.Suggestion[commandinput.CommandMetadata[T]], error),
) {
	updateContext(cmd, "completer", f)
}

func getCompleter[T any](cmd *cobra.Command) *func(cmd *cobra.Command, args []string,
	toComplete string) ([]suggestion.Suggestion[commandinput.CommandMetadata[T]], error) {
	val := contextVal(cmd, "completer")
	if val == nil {
		return nil
	}
	funcVal := val.(func(cmd *cobra.Command, args []string,
		toComplete string) ([]suggestion.Suggestion[commandinput.CommandMetadata[T]], error))
	return &funcVal
}

func GetSelectedSuggestion[T any](cmd *cobra.Command) *suggestion.Suggestion[commandinput.CommandMetadata[T]] {
	val := contextVal(cmd.Root(), "selected")
	if val == nil {
		return nil
	}
	suggestion := val.(*suggestion.Suggestion[commandinput.CommandMetadata[T]])
	return suggestion
}

func setSelectedSuggestion[T any](cmd *cobra.Command,
	selected *suggestion.Suggestion[commandinput.CommandMetadata[T]],
) {
	updateContext(cmd.Root(), "selected", selected)
}

func ShowFlagPlaceholder(cmd *cobra.Command, show bool) {
	updateContext(cmd, "showFlagPlaceholder", show)
}

func getShowFlagPlaceholder(cmd *cobra.Command) *bool {
	val := contextVal(cmd, "showFlagPlaceholder")
	if val == nil {
		return nil
	}
	boolVal := val.(bool)
	return &boolVal
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
