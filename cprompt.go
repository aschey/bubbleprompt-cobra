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

var interactive bool = false

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

	var err error = nil
	cobraCommand := m.rootCmd
	if m.textInput.CommandCompleted() {
		cobraCommand, _, err = m.rootCmd.Find(append([]string{m.textInput.CommandBeforeCursor()}, m.textInput.CompletedArgsBeforeCursor()...))
	}
	if err != nil {
		return nil, err
	}
	text := m.textInput.CurrentTokenBeforeCursor(commandinput.RoundUp)
	tokenPos := m.textInput.CurrentTokenPos(commandinput.RoundUp).Index
	allValues := m.textInput.AllValues()
	prevToken := ""
	if len(allValues) > 0 && tokenPos > 0 && len(allValues) > tokenPos {
		prevToken = allValues[tokenPos-1]
	}

	lastChar := ""
	if m.textInput.Cursor() > len(m.textInput.Value()) {
		lastChar = string(m.textInput.Value()[m.textInput.Cursor()-1])
	}

	var isInMiddle bool

	if strings.HasPrefix(prevToken, "-") {
		isInMiddle = lastChar != " " && lastChar != "="
	} else {
		isInMiddle = lastChar != " "
	}

	if isInMiddle {
		tokenPos++
	}

	if cobraCommand.ValidArgsFunction != nil {
		args, _ := cobraCommand.ValidArgsFunction(cobraCommand, allValues[1:tokenPos], text)

		for _, arg := range args {
			suggestions = append(suggestions, input.Suggestion[cobraMetadata]{
				Text: arg,
				Metadata: cobraMetadata{
					commandinput.CmdMetadata{},
					cobraCommand,
				},
			})
		}
	}
	suggestions = append(suggestions, m.getSubcommandSuggestions(*cobraCommand)...)

	if cobraCommand.Args != nil {
		err = cobraCommand.Args(cobraCommand, m.textInput.ArgsBeforeCursor())
	}

	useParts := strings.Split(cobraCommand.Use, " ")
	placeholders := []string{}
	if len(useParts) > 1 {
		placeholders = useParts[1:]
	}
	placeholdersBeforeFlags := len(placeholders)
	if len(placeholders) > 0 && placeholders[len(placeholders)-1] == "[flags]" {
		placeholdersBeforeFlags--
	}

	if err == nil && (len(m.textInput.ArgsBeforeCursor()) >= placeholdersBeforeFlags || strings.HasPrefix(m.textInput.CurrentTokenBeforeCursor(commandinput.RoundUp), "-")) {
		flags := []commandinput.Flag{}

		cobraCommand.Flags().VisitAll(func(flag *pflag.Flag) {
			if flag.Name == "help" {
				return
			}
			placeholder := ""
			if flag.NoOptDefVal == "" {
				placeholder = flag.Value.Type()
				if strings.HasSuffix(placeholder, "64") || strings.HasSuffix(placeholder, "32") {
					placeholder = placeholder[:len(placeholder)-2]
				}
				placeholder = fmt.Sprintf("<%s>", placeholder)
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

		flagSuggestions := m.textInput.FlagSuggestions(text, flags, func(flag commandinput.Flag) cobraMetadata {
			m := commandinput.CmdMetadata{
				PreservePlaceholder: getPreservePlaceholder(cobraCommand, flag.Long),
				FlagPlaceholder: commandinput.Placeholder{
					Text:  flag.Placeholder,
					Style: input.Text{Style: lipgloss.NewStyle().Foreground(lipgloss.Color("14"))},
				},
			}
			return cobraMetadata{
				m,
				cobraCommand,
			}
		})

		suggestions = append(suggestions, flagSuggestions...)
	}

	return completers.FilterHasPrefix(m.textInput.CurrentTokenBeforeCursor(commandinput.RoundUp), suggestions), nil
}

func (m completerModel) getLevel(command cobra.Command) int {
	level := 0
	for command.HasParent() {
		level++
		command = *command.Parent()
	}
	return level
}

func (m completerModel) getSubcommandSuggestions(command cobra.Command) []input.Suggestion[cobraMetadata] {
	suggestions := []input.Suggestion[cobraMetadata]{}
	level := m.getLevel(command)
	for _, c := range command.Commands() {
		if !slices.Contains(m.ignoreCmds, c.Name()) {
			useParts := strings.Split(c.Use, " ")
			placeholders := []string{}
			if len(useParts) > 1 {
				placeholders = useParts[1:]
			}
			args := []commandinput.PositionalArg{}

			for _, arg := range placeholders {
				args = append(args, commandinput.NewPositionalArg(arg))
			}

			cobraCommand := c
			suggestions = append(suggestions, input.Suggestion[cobraMetadata]{
				Text:        c.Name(),
				Description: c.Short,
				Metadata: cobraMetadata{
					commandinput.CmdMetadata{PositionalArgs: args, Level: level},
					cobraCommand,
				},
			})
		}
	}

	return suggestions
}

func (m completerModel) executor(input string, selectedSuggestion *input.Suggestion[cobraMetadata]) (tea.Model, error) {
	m.rootCmd.SetArgs(m.textInput.AllValues())
	selected := m.textInput.SelectedCommand()
	if selected == nil {
		return executors.NewStringModel(""), fmt.Errorf("No command selected")
	}

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
	if len(out) > 0 {
		return executors.NewStringModel(string(out)), nil
	}

	model := model(m.rootCmd)

	return model, err
}

func (m *Model) SetIgnoreCmds(ignoreCmds ...string) {
	m.completer.ignoreCmds = ignoreCmds
}

func ExecModel(cmd *cobra.Command, model tea.Model) error {
	//interactive := interactive(cmd)
	if interactive {
		setModel(cmd.Root(), model)
		return nil
	} else {
		model, err := tea.NewProgram(model).StartReturningModel()

		fmt.Println(model.View())
		return err
	}
}

func NewPrompt(cmd *cobra.Command) Model {
	interactive = true
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
