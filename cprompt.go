package cprompt

import (
	"fmt"
	"io/ioutil"
	"os"

	prompt "github.com/aschey/bubbleprompt"
	"github.com/aschey/bubbleprompt/input"
	"github.com/aschey/bubbleprompt/input/commandinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

type model struct {
	prompt prompt.Model
}

type completerModel struct {
	suggestions []input.Suggestion
	textInput   *commandinput.Model
}

type executorModel struct {
	rootCmd *cobra.Command
}

func (m model) Init() tea.Cmd {
	return m.prompt.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	p, cmd := m.prompt.Update(msg)
	m.prompt = p
	return m, cmd
}

func (m model) View() string {
	return m.prompt.View()
}

func (m completerModel) completer(document prompt.Document, promptModel prompt.Model) []input.Suggestion {
	return prompt.FilterHasPrefix(m.textInput.CommandBeforeCursor(), m.suggestions)
}

func (m executorModel) executor(input string, selected *input.Suggestion, suggestions []input.Suggestion) tea.Model {

	m.rootCmd.SetArgs([]string{selected.Text})

	rescueStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	m.rootCmd.Execute()
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = rescueStdout

	return prompt.NewStringModel(string(out))
}

func NewPrompt(cmd *cobra.Command) {
	rootCmd := cmd.Root()
	suggestions := []input.Suggestion{}
	for _, c := range rootCmd.Commands() {
		suggestions = append(suggestions, input.Suggestion{Text: c.Name(), Description: c.Short, Metadata: c})
	}

	textInput := commandinput.New()
	completerModel := completerModel{suggestions: suggestions, textInput: textInput}
	executorModel := executorModel{rootCmd: rootCmd}

	m := model{prompt: prompt.New(
		completerModel.completer,
		executorModel.executor,
		textInput,
	)}

	if err := tea.NewProgram(m).Start(); err != nil {
		fmt.Printf("Could not start program :(\n%v\n", err)
		os.Exit(1)
	}
}
