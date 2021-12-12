package apphttp

import "go.uber.org/fx"

type (
	InOptions struct {
		fx.In
		Options []Option `group:"httpOptions"`
	}

	OutOption struct {
		fx.Out
		Option Option `group:"httpOptions"`
	}
)

func FXOptions(options ...Option) fx.Option {
	fxOptions := make([]fx.Option, len(options))
	for i, opt := range options {
		func(i int, opt Option) {
			fxOptions[i] = fx.Provide(func() (out OutOption) {
				out.Option = opt
				return
			})
		}(i, opt)
	}
	return fx.Options(fxOptions...)
}

func FXOption(option Option) fx.Option {
	return FXOptions(option)
}
