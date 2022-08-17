package cprompt

import (
	"fmt"
	"io"
	"os"
	"strings"

	prompt "github.com/aschey/bubbleprompt"
	completers "github.com/aschey/bubbleprompt/completer"
	executors "github.com/aschey/bubbleprompt/executor"
	"github.com/aschey/bubbleprompt/input"
	"github.com/aschey/bubbleprompt/input/commandinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/exp/slices"
)

type Model struct {
	prompt    prompt.Model[cobraMetadata]
	completer completerModel
}

type completerModel struct {
	textInput  *commandinput.Model[cobraMetadata]
	rootCmd    *cobra.Command
	ignoreCmds []string
}

type cobraMetadata struct {
	commandinput.CmdMetadata
	cobraCommand *cobra.Command
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

func (m completerModel) completer(document prompt.Document, promptModel prompt.Model[cobraMetadata]) ([]input.Suggestion[cobraMetadata], error) {
	suggestions := []input.Suggestion[cobraMetadata]{}

	if m.textInput.CommandCompleted() {
		cobraCommand, _, err := m.rootCmd.Find([]string{m.textInput.ParsedValue().Command.Value})
		if err != nil {
			return nil, err
		}
		text := m.textInput.CurrentTokenBeforeCursor(commandinput.RoundUp)
		tokenPos := m.textInput.CurrentTokenPos(commandinput.RoundUp).Index
		allValues := m.textInput.AllValues()
		prevToken := ""
		if len(allValues) > 0 {
			prevToken = allValues[tokenPos-1]
		}

		lastChar := m.textInput.Value()[m.textInput.Cursor()-1]
		var isInMiddle bool

		if strings.HasPrefix(prevToken, "-") {
			isInMiddle = lastChar != ' ' && lastChar != '='
		} else {
			isInMiddle = lastChar != ' '
		}

		if isInMiddle {
			tokenPos++
		}

		posArgs := []commandinput.PositionalArg{}
		if cobraCommand.ValidArgsFunction != nil {
			args, _ := cobraCommand.ValidArgsFunction(cobraCommand, allValues[1:tokenPos], text)
			placeholders := placeholders(cobraCommand)

			for _, posArg := range placeholders {
				posArgs = append(posArgs, commandinput.NewPositionalArg(posArg))
			}

			for _, arg := range args {
				suggestions = append(suggestions, input.Suggestion[cobraMetadata]{
					Text: arg,
					Metadata: cobraMetadata{
						commandinput.NewCmdMetadata(posArgs, commandinput.Placeholder{}),
						cobraCommand,
					},
				})
			}
		}

		if cobraCommand.Args != nil {
			err = cobraCommand.Args(cobraCommand, m.textInput.ArgsBeforeCursor())
		}

		if err == nil && (len(m.textInput.ArgsBeforeCursor()) >= len(posArgs) || strings.HasPrefix(m.textInput.CurrentTokenBeforeCursor(commandinput.RoundUp), "-")) {
			flags := []commandinput.Flag{}
			cobraCommand.Flags().VisitAll(func(flag *pflag.Flag) {
				placeholder := ""
				if flag.NoOptDefVal == "" {
					placeholder = "<" + flag.Value.Type() + ">"
				}
				flags = append(flags, commandinput.Flag{
					Short:            flag.Shorthand,
					Long:             flag.Name,
					RequiresArg:      flag.NoOptDefVal == "",
					Placeholder:      placeholder,
					Description:      flag.Usage,
					PlaceholderStyle: input.Text{Style: lipgloss.NewStyle().Foreground(lipgloss.Color("14"))},
				})
			})

			flagSuggestions := m.textInput.FlagSuggestions(text, flags, func(metadata commandinput.CmdMetadata, flag commandinput.Flag) cobraMetadata {
				m := commandinput.NewCmdMetadata(posArgs, commandinput.Placeholder{
					Text:  flag.Placeholder,
					Style: input.Text{Style: lipgloss.NewStyle().Foreground(lipgloss.Color("14"))}})
				return cobraMetadata{
					m,
					cobraCommand,
				}
			})

			suggestions = append(suggestions, flagSuggestions...)
		}

		return suggestions, nil
	} else {
		for _, c := range m.rootCmd.Commands() {
			if !slices.Contains(m.ignoreCmds, c.Name()) {
				placeholders := placeholders(c)
				args := []commandinput.PositionalArg{}
				flags := []commandinput.Flag{}
				for _, arg := range placeholders {
					args = append(args, commandinput.NewPositionalArg(arg))
				}
				c.Flags().VisitAll(func(flag *pflag.Flag) {
					flags = append(flags, commandinput.Flag{
						Short:       flag.Shorthand,
						Long:        flag.Name,
						Placeholder: flag.Value.Type(),
						RequiresArg: flag.NoOptDefVal == "",
						PlaceholderStyle: input.Text{
							Style: lipgloss.NewStyle().Foreground(lipgloss.Color("14")),
						},
					})
				})
				cobraCommand := c
				suggestions = append(suggestions, input.Suggestion[cobraMetadata]{
					Text:        c.Name(),
					Description: c.Short,
					Metadata: cobraMetadata{
						commandinput.NewCmdMetadata(args, commandinput.Placeholder{}),
						cobraCommand,
					},
				})
			}
		}

		return completers.FilterHasPrefix(m.textInput.CommandBeforeCursor(), suggestions), nil

	}
}

func (m completerModel) executor(input string) (tea.Model, error) {
	m.rootCmd.SetArgs(m.textInput.AllValues())
	selected := m.textInput.SelectedCommand()
	if selected == nil {
		return executors.NewStringModel(""), fmt.Errorf("No command selected")
	}
	setInteractive(selected.Metadata.cobraCommand)

	rescueStdout := os.Stdout
	rescueStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w
	err := m.rootCmd.Execute()
	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = rescueStdout
	os.Stderr = rescueStderr
	model := model(selected.Metadata.cobraCommand)
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
		model, err := tea.NewProgram(model).StartReturningModel()

		fmt.Println(model.View())
		return err
	}
}

func NewPrompt(cmd *cobra.Command) Model {
	rootCmd := cmd.Root()
	curCmd := cmd.Name()

	var textInput input.Input[cobraMetadata] = commandinput.New[cobraMetadata]()
	completerModel := completerModel{
		rootCmd:    rootCmd,
		textInput:  textInput.(*commandinput.Model[cobraMetadata]),
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
