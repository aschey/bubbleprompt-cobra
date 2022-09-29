package cprompt

type Option func(model *Model) error

func WithIgnoreCmds(cmds ...string) Option {
	return func(model *Model) error {
		model.SetIgnoreCmds(cmds...)
		return nil
	}
}

func WithOnCompleterStart(onCompleterStart CompleterStart) Option {
	return func(model *Model) error {
		model.completer.onCompleterStart = onCompleterStart
		return nil
	}
}

func WithOnCompleterFinish(onCompleterStart CompleterFinish) Option {
	return func(model *Model) error {
		model.completer.onCompleterFinish = onCompleterStart
		return nil
	}
}

func WithOnExecutorStart(onExecutorStart ExecutorStart) Option {
	return func(model *Model) error {
		model.completer.onExecutorStart = onExecutorStart
		return nil
	}
}

func WithOnExecutorFinish(onExecutorFinish ExecutorFinish) Option {
	return func(model *Model) error {
		model.completer.onExecutorFinish = onExecutorFinish
		return nil
	}
}
