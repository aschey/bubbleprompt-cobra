package cprompt

import prompt "github.com/aschey/bubbleprompt"

type Option func(model *Model)

func WithIgnoreCmds(cmds ...string) Option {
	return func(model *Model) {
		model.SetIgnoreCmds(cmds...)
	}
}

func WithOnCompleterStart(onCompleterStart CompleterStart) Option {
	return func(model *Model) {
		model.app.onCompleterStart = onCompleterStart
	}
}

func WithOnCompleterFinish(onCompleterStart CompleterFinish) Option {
	return func(model *Model) {
		model.app.onCompleterFinish = onCompleterStart
	}
}

func WithOnExecutorStart(onExecutorStart ExecutorStart) Option {
	return func(model *Model) {
		model.app.onExecutorStart = onExecutorStart
	}
}

func WithOnExecutorFinish(onExecutorFinish ExecutorFinish) Option {
	return func(model *Model) {
		model.app.onExecutorFinish = onExecutorFinish
	}
}

func WithPromptOptions(options ...prompt.Option[CobraMetadata]) Option {
	return func(model *Model) {
		prompt := buildAppModel(*model.app, options...)
		model.prompt = prompt
	}
}
