package cprompt

import (
	"io/ioutil"
	"os"

	prompt "github.com/aschey/bubbleprompt"
	completers "github.com/aschey/bubbleprompt/completer"
	executors "github.com/aschey/bubbleprompt/executor"
	"github.com/aschey/bubbleprompt/input"
	"github.com/aschey/bubbleprompt/input/commandinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
)

type Model struct {
	prompt    prompt.Model
	completer completerModel
}

type completerModel struct {
	textInput  *commandinput.Model
	rootCmd    *cobra.Command
	ignoreCmds []string
}

func (m Model) Init() tea.Cmd {
	return m.prompt.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	p, cmd := m.prompt.Update(msg)
	m.prompt = p
	return m, cmd
}

func (m Model) View() string {
	return m.prompt.View()
}

func (m completerModel) completer(document prompt.Document, promptModel prompt.Model) []input.Suggestion {
	suggestions := []input.Suggestion{}
	for _, c := range m.rootCmd.Commands() {
		cmdName := c.Name()
		if !slices.Contains(m.ignoreCmds, cmdName) {
			suggestions = append(suggestions, input.Suggestion{Text: c.Name(), Description: c.Short, Metadata: c})
		}
	}
	return completers.FilterHasPrefix(m.textInput.CommandBeforeCursor(), suggestions)
}

func (m completerModel) executor(input string, selected *input.Suggestion, suggestions []input.Suggestion) (tea.Model, error) {
	m.rootCmd.SetArgs(m.textInput.AllValues())

	rescueStdout := os.Stdout
	rescueStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w
	err := m.rootCmd.Execute()
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = rescueStdout
	os.Stderr = rescueStderr

	return executors.NewStringModel(string(out)), err
}

func (m *Model) SetIgnoreCmds(ignoreCmds ...string) {
	m.completer.ignoreCmds = ignoreCmds
}

func NewPrompt(cmd *cobra.Command) Model {
	rootCmd := cmd.Root()
	curCmd := cmd.Name()

	textInput := commandinput.New()
	completerModel := completerModel{
		rootCmd: rootCmd,

		textInput:  textInput,
		ignoreCmds: []string{curCmd, "completion"},
	}

	m := Model{
		prompt: prompt.New(
			completerModel.completer,
			completerModel.executor,
			textInput,
		),
	}

	return m
}

func (m Model) Start() error {
	return tea.NewProgram(m).Start()
}
