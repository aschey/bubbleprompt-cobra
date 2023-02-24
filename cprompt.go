package cprompt

import (
	"fmt"
	"io"
	"os"
	"strings"

	prompt "github.com/aschey/bubbleprompt"
	"github.com/aschey/bubbleprompt/completer"
	executors "github.com/aschey/bubbleprompt/executor"
	"github.com/aschey/bubbleprompt/input/commandinput"
	"github.com/aschey/bubbleprompt/suggestion"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/exp/slices"
)

type Model struct {
	prompt prompt.Model[CobraMetadata]
	app    *appModel
}
type CompleterStart func(promptModel prompt.Model[CobraMetadata])

type (
	CompleterFinish func(suggestions []suggestion.Suggestion[CobraMetadata], err error) (
		[]suggestion.Suggestion[CobraMetadata], error)
	ExecutorStart  func(input string, selectedSuggestion *suggestion.Suggestion[CobraMetadata])
	ExecutorFinish func(model tea.Model, err error) (tea.Model, error)
)

type appModel struct {
	textInput         *commandinput.Model[CobraMetadata]
	rootCmd           *cobra.Command
	onCompleterStart  CompleterStart
	onCompleterFinish CompleterFinish
	onExecutorStart   ExecutorStart
	onExecutorFinish  ExecutorFinish
	ignoreCmds        []string
	filterer          completer.Filterer[CobraMetadata]
}

type CobraMetadata struct {
	commandinput.CommandMetadata
	cobraCommand *cobra.Command
}

var interactive bool = false

func (m Model) Init() tea.Cmd {
	return m.prompt.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	p, cmd := m.prompt.Update(msg)
	m.prompt = p.(prompt.Model[CobraMetadata])
	return m, cmd
}

func (m Model) View() string {
	return m.prompt.View()
}

func (m appModel) Complete(
	promptModel prompt.Model[CobraMetadata],
) ([]suggestion.Suggestion[CobraMetadata], error) {
	if m.onCompleterStart != nil {
		m.onCompleterStart(promptModel)
	}
	suggestions := []suggestion.Suggestion[CobraMetadata]{}

	var err error = nil
	cobraCommand := m.rootCmd
	if m.textInput.CommandCompleted() {
		cobraCommand, _, err = m.rootCmd.Find(
			append(
				[]string{m.textInput.CommandBeforeCursor()},
				m.textInput.CompletedArgsBeforeCursor()...),
		)
	}
	if err != nil {
		return nil, err
	}
	text := m.textInput.CurrentTokenBeforeCursor().Value
	level := m.getLevel(*cobraCommand)

	if cobraCommand.ValidArgsFunction != nil {
		suggestions = append(suggestions, m.getValidArgSuggestions(cobraCommand, level)...)
	}
	subcommandSuggestions, err := m.getSubcommandSuggestions(*cobraCommand)
	if err != nil {
		return nil, err
	}
	suggestions = append(suggestions, subcommandSuggestions...)

	placeholderStr := usageArgs(cobraCommand.Use)
	placeholders, err := m.textInput.ParseUsage(placeholderStr)
	if err != nil {
		return nil, err
	}
	placeholdersBeforeFlags := len(placeholders)
	if len(placeholders) > 0 && placeholders[len(placeholders)-1].Placeholder() == "[flags]" {
		placeholdersBeforeFlags--
	}
	argsBeforeCursor := m.textInput.ArgsBeforeCursor()
	flags := m.textInput.ParsedValue().Flags

	// Always show flag suggestions if the user already entered a flag
	if err == nil &&
		(len(argsBeforeCursor)-level >= placeholdersBeforeFlags ||
			strings.HasPrefix(m.textInput.CurrentTokenBeforeCursor().Value, "-") ||
			len(flags) > 0) {
		flagSuggestions := m.getFLagSuggestions(text, cobraCommand)

		if len(flagSuggestions) > 0 {
			return flagSuggestions, nil
		}
	}

	result := m.filterer.Filter(m.textInput.CurrentTokenBeforeCursor().Value, suggestions)
	if m.onCompleterFinish != nil {
		return m.onCompleterFinish(result, nil)
	}
	return result, nil
}

func (m appModel) getFLagSuggestions(text string, cobraCommand *cobra.Command) []suggestion.Suggestion[CobraMetadata] {
	flags := []commandinput.FlagInput{}

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
		flags = append(flags, commandinput.FlagInput{
			Short:          flag.Shorthand,
			Long:           flag.Name,
			ArgPlaceholder: m.textInput.NewFlagPlaceholder(placeholder),
			Description:    flag.Usage,
		})
	})

	return m.textInput.FlagSuggestions(
		text,
		flags,
		func(flag commandinput.FlagInput) CobraMetadata {
			m := commandinput.CommandMetadata{
				PreservePlaceholder: getPreservePlaceholder(cobraCommand, flag.Long),
				FlagArgPlaceholder:  flag.ArgPlaceholder,
			}
			return CobraMetadata{
				m,
				cobraCommand,
			}
		},
	)
}

func (m appModel) getValidArgSuggestions(
	cobraCommand *cobra.Command, level int,
) []suggestion.Suggestion[CobraMetadata] {
	completed := m.textInput.CompletedArgsBeforeCursor()[level:]
	validArgs, _ := cobraCommand.ValidArgsFunction(
		cobraCommand,
		completed,
		m.textInput.CurrentTokenBeforeCursorRoundDown().Value,
	)
	suggestions := []suggestion.Suggestion[CobraMetadata]{}
	for _, arg := range validArgs {
		suggestions = append(suggestions, suggestion.Suggestion[CobraMetadata]{
			Text: arg,
			Metadata: CobraMetadata{
				commandinput.CommandMetadata{
					ShowFlagPlaceholder: hasUserDefinedFlags(cobraCommand),
				},
				cobraCommand,
			},
		})
	}

	return suggestions
}

func (m appModel) getLevel(command cobra.Command) int {
	level := 0
	for command.HasParent() {
		level++
		command = *command.Parent()
	}
	return level - 1
}

func (m appModel) getSubcommandSuggestions(
	command cobra.Command,
) ([]suggestion.Suggestion[CobraMetadata], error) {
	suggestions := []suggestion.Suggestion[CobraMetadata]{}
	for _, c := range command.Commands() {
		if !slices.Contains(m.ignoreCmds, c.Name()) {
			placeholders := usageArgs(c.Use)
			args, err := m.textInput.ParseUsage(placeholders)
			if err != nil {
				return nil, err
			}

			cobraCommand := c
			hasFlags := hasUserDefinedFlags(c)

			if len(args) > 0 && args[len(args)-1].Placeholder() == "[flags]" {
				hasFlags = false
			}
			suggestions = append(suggestions, suggestion.Suggestion[CobraMetadata]{
				Text:        c.Name(),
				Description: c.Short,
				Metadata: CobraMetadata{
					commandinput.CommandMetadata{
						PositionalArgs:      args,
						ShowFlagPlaceholder: hasFlags,
					},
					cobraCommand,
				},
			})
		}
	}

	return suggestions, nil
}

func (m appModel) Update(msg tea.Msg) (prompt.InputHandler[CobraMetadata], tea.Cmd) {
	return m, nil
}

func (m appModel) Execute(
	input string,
	promptModel *prompt.Model[CobraMetadata],
) (tea.Model, error) {
	if m.onExecutorStart != nil {
		m.onExecutorStart(input, promptModel.SuggestionManager().SelectedSuggestion())
	}
	all := m.textInput.Values()
	if len(all[0]) == 0 {
		err := fmt.Errorf("No command selected")
		if m.onExecutorFinish != nil {
			return m.onExecutorFinish(nil, err)
		}
		return nil, err
	}

	m.rootCmd.SetArgs(all)

	// Reset flags before each run to ensure old values are cleared out
	cmd, _, _ := m.rootCmd.Find(all)
	if cmd == nil || !cmd.Runnable() {
		return nil, fmt.Errorf("Invalid command")
	}

	if cmd != nil {
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			_ = f.Value.Set(f.DefValue)
		})
	}

	rescueStdout := os.Stdout
	rescueStderr := os.Stderr
	outR, outW, _ := os.Pipe()
	errR, errW, _ := os.Pipe()
	os.Stdout = outW
	os.Stderr = errW

	err := m.rootCmd.Execute()
	if outErr := outW.Close(); outErr != nil {
		return nil, outErr
	}
	if outErr := errW.Close(); err != nil {
		return nil, outErr
	}
	outData, outErr := io.ReadAll(outR)
	if outErr != nil {
		return nil, outErr
	}
	errData, outErr := io.ReadAll(errR)
	if outErr != nil {
		return nil, outErr
	}
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
	m.app.ignoreCmds = ignoreCmds
}

func (m *Model) SetOnCompleterStart(onCompleterStart CompleterStart) {
	m.app.onCompleterStart = onCompleterStart
}

func (m *Model) SetOnCompleterFinish(onCompleterFinish CompleterFinish) {
	m.app.onCompleterFinish = onCompleterFinish
}

func (m *Model) SetOnExecutorStart(onExecutorStart ExecutorStart) {
	m.app.onExecutorStart = onExecutorStart
}

func (m *Model) SetOnExecutorFinish(onExecutorFinish ExecutorFinish) {
	m.app.onExecutorFinish = onExecutorFinish
}

func (m *Model) SetFilterer(filterer completer.Filterer[CobraMetadata]) {
	m.app.filterer = filterer
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
	return FilterShellCompletionsWith(options, toComplete, completer.NewPrefixFilter[CobraMetadata]())
}

func FilterShellCompletionsWith(options []string, toComplete string,
	filterer completer.Filterer[CobraMetadata],
) []string {
	suggestions := []suggestion.Suggestion[CobraMetadata]{}
	for _, option := range options {
		suggestions = append(suggestions, suggestion.Suggestion[CobraMetadata]{Text: option})
	}
	filtered := filterer.Filter(toComplete, suggestions)
	results := []string{}
	for _, result := range filtered {
		results = append(results, result.Text)
	}
	return results
}

func buildAppModel(app appModel, opts ...prompt.Option[CobraMetadata]) prompt.Model[CobraMetadata] {
	return prompt.New[CobraMetadata](
		app,
		app.textInput,
		opts...,
	)
}

func hasUserDefinedFlags(command *cobra.Command) bool {
	show := getShowFlagPlaceholder(command)
	if show != nil {
		return *show
	}

	hasFlags := false
	command.LocalFlags().VisitAll(func(f *pflag.Flag) {
		if f.Name != "help" {
			hasFlags = true
		}
	})
	return hasFlags
}

func NewPrompt(cmd *cobra.Command, options ...Option) Model {
	interactive = true
	rootCmd := cmd.Root()
	// Don't need usage messages popping up in the prompt, it just adds noise
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true

	curCmd := cmd.Name()

	textInput := commandinput.New[CobraMetadata]()
	app := appModel{
		rootCmd:    rootCmd,
		textInput:  textInput,
		ignoreCmds: []string{curCmd, "completion", "help"},
	}
	prompt := buildAppModel(app)

	m := Model{
		prompt: prompt,
		app:    &app,
	}

	for _, option := range options {
		option(&m)
	}

	return m
}

func usageArgs(s string) string {
	stringParts := 2
	parts := strings.SplitN(s, " ", stringParts)
	if len(parts) < stringParts {
		return ""
	}
	return parts[1]
}

func (m Model) Start() error {
	_, err := tea.NewProgram(m, tea.WithFilter(prompt.MsgFilter)).Run()
	return err
}
