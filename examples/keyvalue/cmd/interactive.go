/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"examples/keyvalue/db"
	"time"

	prompt "github.com/aschey/bubbleprompt"
	cprompt "github.com/aschey/bubbleprompt-cobra"
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
		promptModel, err := cprompt.NewPrompt(cmd)
		if err != nil {
			return err
		}

		model := model{inner: promptModel}
		return tea.NewProgram(&model).Start()
	},
}

type model struct {
	inner cprompt.Model
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.inner.Init(), prompt.PeriodicCompleter(time.Second))
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(prompt.PeriodicCompleterMsg); ok {
		_ = db.LoadDb()
	}
	model, cmd := m.inner.Update(msg)
	m.inner = model.(cprompt.Model)
	return m, cmd
}

func (m model) View() string {
	return m.inner.View()
}

func init() {
	rootCmd.AddCommand(interactiveCmd)
}
