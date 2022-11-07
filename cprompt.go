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
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/exp/slices"
)

type Model struct {
	prompt    prompt.Model[CobraMetadata]
	completer *completerModel
}
type CompleterStart func(promptModel prompt.Model[CobraMetadata])
type CompleterFinish func(suggestions []input.Suggestion[CobraMetadata], err error) ([]input.Suggestion[CobraMetadata], error)
type ExecutorStart func(input string, selectedSuggestion *input.Suggestion[CobraMetadata])
type ExecutorFinish func(model tea.Model, err error) (tea.Model, error)

type completerModel struct {
	textInput         *commandinput.Model[CobraMetadata]
	rootCmd           *cobra.Command
	onCompleterStart  CompleterStart
	onCompleterFinish CompleterFinish
	onExecutorStart   ExecutorStart
	onExecutorFinish  ExecutorFinish
	ignoreCmds        []string
}

type CobraMetadata struct {
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

func (m *completerModel) completer(promptModel prompt.Model[CobraMetadata]) ([]input.Suggestion[CobraMetadata], error) {
	if m.onCompleterStart != nil {
		m.onCompleterStart(promptModel)
	}
	suggestions := []input.Suggestion[CobraMetadata]{}

	var err error = nil
	cobraCommand := m.rootCmd
	if m.textInput.CommandCompleted() {
		cobraCommand, _, err = m.rootCmd.Find(append([]string{m.textInput.CommandBeforeCursor()}, m.textInput.CompletedArgsBeforeCursor()...))
	}
	if err != nil {
		return nil, err
	}
	text := m.textInput.CurrentTokenBeforeCursor(commandinput.RoundUp)
	level := m.getLevel(*cobraCommand)

	if cobraCommand.ValidArgsFunction != nil {
		completed := m.textInput.CompletedArgsBeforeCursor()[level:]
		validArgs, _ := cobraCommand.ValidArgsFunction(cobraCommand, completed, m.textInput.CurrentTokenBeforeCursor(commandinput.RoundDown))

		for _, arg := range validArgs {
			suggestions = append(suggestions, input.Suggestion[CobraMetadata]{
				Text: arg,
				Metadata: CobraMetadata{
					commandinput.CmdMetadata{ShowFlagPlaceholder: hasUserDefinedFlags(cobraCommand)},
					cobraCommand,
				},
			})
		}
	}
	suggestions = append(suggestions, m.getSubcommandSuggestions(*cobraCommand)...)

	useParts := strings.Split(cobraCommand.Use, " ")
	placeholders := []string{}
	if len(useParts) > 1 {
		placeholders = useParts[1:]
	}
	placeholdersBeforeFlags := len(placeholders)
	if len(placeholders) > 0 && placeholders[len(placeholders)-1] == "[flags]" {
		placeholdersBeforeFlags--
	}
	argsBeforeCursor := m.textInput.ArgsBeforeCursor()

	if err == nil && (len(argsBeforeCursor)-level >= placeholdersBeforeFlags || strings.HasPrefix(m.textInput.CurrentTokenBeforeCursor(commandinput.RoundUp), "-")) {
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
				Short:       flag.Shorthand,
				Long:        flag.Name,
				RequiresArg: flag.NoOptDefVal == "",
				Placeholder: m.textInput.NewFlagPlaceholder(placeholder),
				Description: flag.Usage,
			})
		})

		flagSuggestions := m.textInput.FlagSuggestions(text, flags, func(flag commandinput.Flag) CobraMetadata {
			m := commandinput.CmdMetadata{
				PreservePlaceholder: getPreservePlaceholder(cobraCommand, flag.Long),
				FlagPlaceholder:     flag.Placeholder,
			}
			return CobraMetadata{
				m,
				cobraCommand,
			}
		})

		if len(flagSuggestions) > 0 {
			return flagSuggestions, nil
		}
	}

	result := completers.FilterHasPrefix(m.textInput.CurrentTokenBeforeCursor(commandinput.RoundUp), suggestions)
	if m.onCompleterFinish != nil {
		return m.onCompleterFinish(result, nil)
	}
	return result, nil
}

func (m *completerModel) getLevel(command cobra.Command) int {
	level := 0
	for command.HasParent() {
		level++
		command = *command.Parent()
	}
	return level - 1
}

func (m *completerModel) getSubcommandSuggestions(command cobra.Command) []input.Suggestion[CobraMetadata] {
	suggestions := []input.Suggestion[CobraMetadata]{}
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
				args = append(args, m.textInput.NewPositionalArg(arg))
			}

			cobraCommand := c
			hasFlags := hasUserDefinedFlags(c)

			if len(args) > 0 && args[len(args)-1].Placeholder() == "[flags]" {
				hasFlags = false
			}
			suggestions = append(suggestions, input.Suggestion[CobraMetadata]{
				Text:        c.Name(),
				Description: c.Short,
				Metadata: CobraMetadata{
					commandinput.CmdMetadata{PositionalArgs: args, Level: level + 1, ShowFlagPlaceholder: hasFlags},
					cobraCommand,
				},
			})
		}
	}

	return suggestions
}

func (m *completerModel) executor(input string, selectedSuggestion *input.Suggestion[CobraMetadata]) (tea.Model, error) {
	if m.onExecutorStart != nil {
		m.onExecutorStart(input, selectedSuggestion)
	}
	m.rootCmd.SetArgs(m.textInput.AllValues())

	// Reset flags before each run to ensure old values are cleared out
	cmd, _, _ := m.rootCmd.Find(m.textInput.AllValues())
	if cmd != nil {
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			_ = f.Value.Set(f.DefValue)
		})
	}

	selected := m.textInput.SelectedCommand()
	if selected == nil {
		return nil, fmt.Errorf("No command selected")
	}

	rescueStdout := os.Stdout
	rescueStderr := os.Stderr
	outR, outW, _ := os.Pipe()
	errR, errW, _ := os.Pipe()
	os.Stdout = outW
	os.Stderr = errW

	err := m.rootCmd.Execute()
	outW.Close()
	errW.Close()
	outData, _ := io.ReadAll(outR)
	errData, _ := io.ReadAll(errR)
	os.Stdout = rescueStdout
	os.Stderr = rescueStderr
	if len(outData) > 0 {
		return executors.NewStringModel(string(outData)), nil
	}
	if len(errData) > 0 {
		return nil, fmt.Errorf(strings.TrimRight(string(errData), "\n"))
	}

	model := model(m.rootCmd)

	if m.onExecutorFinish != nil {
		return m.onExecutorFinish(model, err)
	}
	return model, err
}

func (m *Model) SetIgnoreCmds(ignoreCmds ...string) {
	m.completer.ignoreCmds = ignoreCmds
}

func (m *Model) SetOnCompleterStart(onCompleterStart CompleterStart) {
	m.completer.onCompleterStart = onCompleterStart
}

func (m *Model) SetOnCompleterFinish(onCompleterFinish CompleterFinish) {
	m.completer.onCompleterFinish = onCompleterFinish
}

func (m *Model) SetOnExecutorStart(onExecutorStart ExecutorStart) {
	m.completer.onExecutorStart = onExecutorStart
}

func (m *Model) SetOnExecutorFinish(onExecutorFinish ExecutorFinish) {
	m.completer.onExecutorFinish = onExecutorFinish
}

func ExecModel(cmd *cobra.Command, model tea.Model) error {
	if interactive {
		setModel(cmd.Root(), model)
		return nil
	} else {
		_, err := tea.NewProgram(model).Run()
		return err
	}
}

func FilterShellCompletions(options []string, toComplete string) []string {
	suggestions := []input.Suggestion[CobraMetadata]{}
	for _, option := range options {
		suggestions = append(suggestions, input.Suggestion[CobraMetadata]{Text: option})
	}
	filtered := completers.FilterHasPrefix(toComplete, suggestions)
	results := []string{}
	for _, result := range filtered {
		results = append(results, result.Text)
	}
	return results
}

func buildPromptModel(completerModel completerModel, opts ...prompt.Option[CobraMetadata]) (prompt.Model[CobraMetadata], error) {
	return prompt.New(
		completerModel.completer,
		completerModel.executor,
		completerModel.textInput,
		opts...,
	)
}

func hasUserDefinedFlags(command *cobra.Command) bool {
	hasFlags := false
	command.LocalFlags().VisitAll(func(f *pflag.Flag) {
		if f.Name != "help" {
			hasFlags = true
		}
	})
	return hasFlags
}

func NewPrompt(cmd *cobra.Command, options ...Option) (Model, error) {
	interactive = true
	rootCmd := cmd.Root()
	// Don't need usage messages popping up in the prompt, it just adds noise
	rootCmd.SilenceUsage = true
	curCmd := cmd.Name()

	textInput := commandinput.New[CobraMetadata]()
	completerModel := completerModel{
		rootCmd:    rootCmd,
		textInput:  textInput,
		ignoreCmds: []string{curCmd, "completion", "help"},
	}
	prompt, err := buildPromptModel(completerModel)
	if err != nil {
		return Model{}, err
	}

	m := Model{
		prompt:    prompt,
		completer: &completerModel,
	}

	for _, option := range options {
		if err := option(&m); err != nil {
			return Model{}, err
		}
	}

	return m, nil
}

func (m Model) Start() error {
	_, err := tea.NewProgram(m, tea.WithFilter(prompt.MsgFilter)).Run()
	return err
}
