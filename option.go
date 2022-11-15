package cprompt

import prompt "github.com/aschey/bubbleprompt"

type Option func(model *Model) error

func WithIgnoreCmds(cmds ...string) Option {
	return func(model *Model) error {
		model.SetIgnoreCmds(cmds...)
		return nil
	}
}

func WithOnCompleterStart(onCompleterStart CompleterStart) Option {
	return func(model *Model) error {
		model.app.onCompleterStart = onCompleterStart
		return nil
	}
}

func WithOnCompleterFinish(onCompleterStart CompleterFinish) Option {
	return func(model *Model) error {
		model.app.onCompleterFinish = onCompleterStart
		return nil
	}
}

func WithOnExecutorStart(onExecutorStart ExecutorStart) Option {
	return func(model *Model) error {
		model.app.onExecutorStart = onExecutorStart
		return nil
	}
}

func WithOnExecutorFinish(onExecutorFinish ExecutorFinish) Option {
	return func(model *Model) error {
		model.app.onExecutorFinish = onExecutorFinish
		return nil
	}
}

func WithPromptOptions(options ...prompt.Option[CobraMetadata]) Option {
	return func(model *Model) error {
		prompt, err := buildAppModel(*model.app, options...)
		if err != nil {
			return err
		}
		model.prompt = prompt
		return nil
	}
}
