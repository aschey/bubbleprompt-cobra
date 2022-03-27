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

	if !m.textInput.CommandCompleted() {
		for _, c := range m.rootCmd.Commands() {
			if !slices.Contains(m.ignoreCmds, c.Name()) {
				placeholders := placeholders(c)
				args := []input.PositionalArg{}
				for _, arg := range placeholders {
					args = append(args, input.NewPositionalArg(arg))
				}
				suggestions = append(suggestions, input.Suggestion{
					Text:           c.Name(),
					Description:    c.Short,
					PositionalArgs: args,
					Metadata:       c,
				})
			}
		}

		return completers.FilterHasPrefix(m.textInput.CommandBeforeCursor(), suggestions)
	} else {
		cmd, _, _ := m.rootCmd.Find([]string{m.textInput.ParsedValue().Command.Value})
		args, _ := cmd.ValidArgsFunction(cmd, m.textInput.AllValues()[1:], m.textInput.CurrentTokenBeforeCursor())
		for _, arg := range args {
			suggestions = append(suggestions, input.Suggestion{
				Text:     arg,
				Metadata: cmd,
			})
		}

		return suggestions
	}
}

func (m completerModel) executor(input string, selected *input.Suggestion, suggestions []input.Suggestion) (tea.Model, error) {
	m.rootCmd.SetArgs(m.textInput.AllValues())
	cmd := selected.Metadata.(*cobra.Command)
	setInteractive(cmd)

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
	model := model(cmd)
	if model == nil {
		return executors.NewStringModel(string(out)), err
	}
	return model, err
}

func (m *Model) SetIgnoreCmds(ignoreCmds ...string) {
	m.completer.ignoreCmds = ignoreCmds
}

func ExecModel(cmd *cobra.Command, model tea.Model) error {
	interactive := interactive(cmd)
	if interactive {
		setModel(cmd, model)
		return nil
	} else {
		return tea.NewProgram(model).Start()
	}
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
