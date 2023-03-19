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

type Model[T any] struct {
	prompt prompt.Model[commandinput.CommandMetadata[T]]
	app    *appModel[T]
}
type CompleterStart[T any] func(promptModel prompt.Model[commandinput.CommandMetadata[T]])

type (
	CompleterFinish[T any] func(suggestions []suggestion.Suggestion[commandinput.CommandMetadata[T]], err error) (
		[]suggestion.Suggestion[commandinput.CommandMetadata[T]], error)
	ExecutorStart[T any] func(input string, selectedSuggestion *suggestion.Suggestion[commandinput.CommandMetadata[T]])
	ExecutorFinish       func(model tea.Model, err error) (tea.Model, error)
)

type appModel[T any] struct {
	textInput         *commandinput.Model[T]
	rootCmd           *cobra.Command
	onCompleterStart  CompleterStart[T]
	onCompleterFinish CompleterFinish[T]
	onExecutorStart   ExecutorStart[T]
	onExecutorFinish  ExecutorFinish
	ignoreCmds        []string
	filterer          completer.Filterer[commandinput.CommandMetadata[T]]
}

var interactive bool = false

func (m Model[T]) Init() tea.Cmd {
	return m.prompt.Init()
}

func (m Model[T]) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	p, cmd := m.prompt.Update(msg)
	m.prompt = p.(prompt.Model[commandinput.CommandMetadata[T]])
	return m, cmd
}

func (m Model[T]) View() string {
	return m.prompt.View()
}

func (m appModel[T]) Complete(
	promptModel prompt.Model[commandinput.CommandMetadata[T]],
) ([]suggestion.Suggestion[commandinput.CommandMetadata[T]], error) {
	if m.onCompleterStart != nil {
		m.onCompleterStart(promptModel)
	}

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

	level := m.getLevel(*cobraCommand)

	suggestions, err := m.getArgSuggestions(cobraCommand, level)
	if err != nil {
		return nil, err
	}

	if len(suggestions) > 0 {
		if m.onCompleterFinish != nil {
			return m.onCompleterFinish(suggestions, nil)
		}
		return suggestions, nil
	}

	subcommandSuggestions, err := m.getSubcommandSuggestions(*cobraCommand)
	if err != nil {
		return nil, err
	}
	suggestions = append(suggestions, subcommandSuggestions...)

	flagSuggestions, err := m.getFlagSuggestions(cobraCommand, level)
	if err != nil {
		return nil, err
	}
	if len(flagSuggestions) > 0 {
		return flagSuggestions, nil
	}

	result := m.filterer.Filter(m.textInput.CurrentTokenBeforeCursor().Value, suggestions)
	if m.onCompleterFinish != nil {
		return m.onCompleterFinish(result, nil)
	}
	return result, nil
}

func (m appModel[T]) getArgSuggestions(cobraCommand *cobra.Command, level int) (
	[]suggestion.Suggestion[commandinput.CommandMetadata[T]], error,
) {
	suggestions := []suggestion.Suggestion[commandinput.CommandMetadata[T]]{}
	completer := getCompleter[T](cobraCommand)
	if completer != nil {
		completed := m.textInput.CompletedArgsBeforeCursor()[level:]
		completerSuggestions, err := (*completer)(
			cobraCommand,
			completed,
			m.textInput.CurrentTokenBeforeCursorRoundDown().Value,
		)
		if err != nil {
			return nil, err
		}
		suggestions = append(suggestions, completerSuggestions...)
	} else if cobraCommand.ValidArgsFunction != nil {
		suggestions = append(suggestions, m.getValidArgSuggestions(cobraCommand, level)...)
	}
	return suggestions, nil
}

func (m appModel[T]) getFlagSuggestions(cobraCommand *cobra.Command, level int) (
	[]suggestion.Suggestion[commandinput.CommandMetadata[T]], error,
) {
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
	if len(argsBeforeCursor)-level >= placeholdersBeforeFlags ||
		strings.HasPrefix(m.textInput.CurrentTokenBeforeCursor().Value, "-") ||
		len(flags) > 0 {
		text := m.textInput.CurrentTokenBeforeCursor().Value
		return m.findFLagSuggestions(text, cobraCommand), nil

	}
	return []suggestion.Suggestion[commandinput.CommandMetadata[T]]{}, nil
}

func (m appModel[T]) findFLagSuggestions(text string, cobraCommand *cobra.Command,
) []suggestion.Suggestion[commandinput.CommandMetadata[T]] {
	flags := []commandinput.FlagInput{}
	if !cobraCommand.HasParent() {
		return []suggestion.Suggestion[commandinput.CommandMetadata[T]]{}
	}
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
		func(flag commandinput.FlagInput) commandinput.CommandMetadata[T] {
			m := commandinput.CommandMetadata[T]{
				PreservePlaceholder: getPreservePlaceholder(cobraCommand, flag.Long),
				FlagArgPlaceholder:  flag.ArgPlaceholder,
			}
			return m
		},
	)
}

func (m appModel[T]) getValidArgSuggestions(
	cobraCommand *cobra.Command, level int,
) []suggestion.Suggestion[commandinput.CommandMetadata[T]] {
	completed := m.textInput.CompletedArgsBeforeCursor()[level:]
	validArgs, _ := cobraCommand.ValidArgsFunction(
		cobraCommand,
		completed,
		m.textInput.CurrentTokenBeforeCursorRoundDown().Value,
	)
	suggestions := []suggestion.Suggestion[commandinput.CommandMetadata[T]]{}
	for _, arg := range validArgs {
		suggestions = append(suggestions, suggestion.Suggestion[commandinput.CommandMetadata[T]]{
			Text: arg,
			Metadata: commandinput.CommandMetadata[T]{
				ShowFlagPlaceholder: hasUserDefinedFlags(cobraCommand),
			},
		})
	}

	return suggestions
}

func (m appModel[T]) getLevel(command cobra.Command) int {
	level := 0
	for command.HasParent() {
		level++
		command = *command.Parent()
	}
	return level - 1
}

func (m appModel[T]) getSubcommandSuggestions(
	command cobra.Command,
) ([]suggestion.Suggestion[commandinput.CommandMetadata[T]], error) {
	suggestions := []suggestion.Suggestion[commandinput.CommandMetadata[T]]{}
	for _, c := range command.Commands() {
		if !slices.Contains(m.ignoreCmds, c.Name()) {
			placeholders := usageArgs(c.Use)
			args, err := m.textInput.ParseUsage(placeholders)
			if err != nil {
				return nil, err
			}

			hasFlags := hasUserDefinedFlags(c)

			if len(args) > 0 && args[len(args)-1].Placeholder() == "[flags]" {
				hasFlags = false
			}
			suggestions = append(suggestions, suggestion.Suggestion[commandinput.CommandMetadata[T]]{
				Text:        c.Name(),
				Description: c.Short,
				Metadata: commandinput.CommandMetadata[T]{
					PositionalArgs:      args,
					ShowFlagPlaceholder: hasFlags,
				},
			})
		}
	}

	return suggestions, nil
}

func (m appModel[T]) Init() tea.Cmd {
	return nil
}

func (m appModel[T]) Update(msg tea.Msg) (prompt.InputHandler[commandinput.CommandMetadata[T]], tea.Cmd) {
	return m, nil
}

func (m appModel[T]) Execute(
	input string,
	promptModel *prompt.Model[commandinput.CommandMetadata[T]],
) (tea.Model, error) {
	setSelectedSuggestion(m.rootCmd, promptModel.SuggestionManager().SelectedSuggestion())
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

func (m *Model[T]) SetIgnoreCmds(ignoreCmds ...string) {
	m.app.ignoreCmds = ignoreCmds
}

func (m *Model[T]) SetOnCompleterStart(onCompleterStart CompleterStart[T]) {
	m.app.onCompleterStart = onCompleterStart
}

func (m *Model[T]) SetOnCompleterFinish(onCompleterFinish CompleterFinish[T]) {
	m.app.onCompleterFinish = onCompleterFinish
}

func (m *Model[T]) SetOnExecutorStart(onExecutorStart ExecutorStart[T]) {
	m.app.onExecutorStart = onExecutorStart
}

func (m *Model[T]) SetOnExecutorFinish(onExecutorFinish ExecutorFinish) {
	m.app.onExecutorFinish = onExecutorFinish
}

func (m *Model[T]) SetFilterer(filterer completer.Filterer[commandinput.CommandMetadata[T]]) {
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

func FilterShellCompletions[T any](options []string, toComplete string) []string {
	return FilterShellCompletionsWith[T](options, toComplete, completer.NewPrefixFilter[commandinput.CommandMetadata[T]]())
}

func FilterShellCompletionsWith[T any](options []string, toComplete string,
	filterer completer.Filterer[commandinput.CommandMetadata[T]],
) []string {
	suggestions := []suggestion.Suggestion[commandinput.CommandMetadata[T]]{}
	for _, option := range options {
		suggestions = append(suggestions, suggestion.Suggestion[commandinput.CommandMetadata[T]]{Text: option})
	}
	filtered := filterer.Filter(toComplete, suggestions)
	results := []string{}
	for _, result := range filtered {
		results = append(results, result.Text)
	}
	return results
}

func buildAppModel[T any](app appModel[T], opts ...prompt.Option[commandinput.CommandMetadata[T]],
) prompt.Model[commandinput.CommandMetadata[T]] {
	return prompt.New[commandinput.CommandMetadata[T]](
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

func NewPrompt[T any](cmd *cobra.Command, options ...Option[T]) Model[T] {
	interactive = true
	rootCmd := cmd.Root()
	// Don't need usage messages popping up in the prompt, it just adds noise
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true

	curCmd := cmd.Name()

	textInput := commandinput.New[T]()
	app := appModel[T]{
		rootCmd:    rootCmd,
		textInput:  textInput,
		filterer:   completer.NewPrefixFilter[commandinput.CommandMetadata[T]](),
		ignoreCmds: []string{curCmd, "completion", "help"},
	}
	prompt := buildAppModel(app)

	m := Model[T]{
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

func (m Model[T]) Start() error {
	_, err := tea.NewProgram(m, tea.WithFilter(prompt.MsgFilter)).Run()
	return err
}
