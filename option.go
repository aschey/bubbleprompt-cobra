package cprompt

import (
	prompt "github.com/aschey/bubbleprompt"
	"github.com/aschey/bubbleprompt/input/commandinput"
)

type Option[T any] func(model *Model[T])

func WithIgnoreCmds[T any](cmds ...string) Option[T] {
	return func(model *Model[T]) {
		model.SetIgnoreCmds(cmds...)
	}
}

func WithOnCompleterStart[T any](onCompleterStart CompleterStart[T]) Option[T] {
	return func(model *Model[T]) {
		model.app.onCompleterStart = onCompleterStart
	}
}

func WithOnCompleterFinish[T any](onCompleterStart CompleterFinish[T]) Option[T] {
	return func(model *Model[T]) {
		model.app.onCompleterFinish = onCompleterStart
	}
}

func WithOnExecutorStart[T any](onExecutorStart ExecutorStart[T]) Option[T] {
	return func(model *Model[T]) {
		model.app.onExecutorStart = onExecutorStart
	}
}

func WithOnExecutorFinish[T any](onExecutorFinish ExecutorFinish) Option[T] {
	return func(model *Model[T]) {
		model.app.onExecutorFinish = onExecutorFinish
	}
}

func WithPromptOptions[T any](options ...prompt.Option[commandinput.CommandMetadata[T]]) Option[T] {
	return func(model *Model[T]) {
		prompt := buildAppModel(*model.app, options...)
		model.prompt = prompt
	}
}
