/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"time"

	"examples/keyvalue/db"

	prompt "github.com/aschey/bubbleprompt"
	cprompt "github.com/aschey/bubbleprompt-cobra"
	"github.com/aschey/bubbleprompt/suggestion"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

// interactiveCmd represents the interactive command
var interactiveCmd = &cobra.Command{
	Use:   "interactive",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		promptModel := cprompt.NewPrompt[any](cmd)

		model := model{inner: promptModel}
		_, err := tea.NewProgram(&model, tea.WithFilter(prompt.MsgFilter)).Run()
		return err
	},
}

type model struct {
	inner cprompt.Model[any]
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.inner.Init(), suggestion.PeriodicCompleter(time.Second))
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(suggestion.PeriodicCompleterMsg); ok {
		_ = db.LoadDb()
	}
	model, cmd := m.inner.Update(msg)
	m.inner = model.(cprompt.Model[any])
	return m, cmd
}

func (m model) View() string {
	return m.inner.View()
}

func init() {
	rootCmd.AddCommand(interactiveCmd)
}
