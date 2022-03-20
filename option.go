package cprompt

type Option func(model *Model) error

func WithIgnoreCmds(cmds ...string) Option {
	return func(model *Model) error {
		model.SetIgnoreCmds(cmds...)
		return nil
	}
}
