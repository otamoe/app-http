package apphttp

import (
	"crypto/tls"
	"log"
	"net/http"
	"time"

	"go.uber.org/fx"
)

type (
	Options struct {
		Addr                   string
		ErrorLog               *log.Logger
		ReadTimeout            time.Duration
		ReadHeaderTimeout      time.Duration
		WriteTimeout           time.Duration
		IdleTimeout            time.Duration
		RequestTimeout         time.Duration
		ShutdownTimeout        time.Duration
		ShutdownRequestTimeout time.Duration
		MaxHeaderBytes         int
		TLSConfig              *tls.Config
		Handlers               []HandlerOption
	}
	InOptions struct {
		fx.In
		Options []Option `group:"httpOptions"`
	}

	OutOption struct {
		fx.Out
		Option Option `group:"httpOptions"`
	}
	Option func(options *Options) error

	HandlerFunc func(next http.Handler) http.Handler

	HandlerOption struct {
		Index   int
		Hosts   []string
		Handler HandlerFunc
	}
)

func Addr(addr string) fx.Option {
	return fx.Provide(func() (out OutOption) {
		out.Option = func(options *Options) error {
			options.Addr = addr
			return nil
		}
		return
	})
}
func ErrorLog(logger *log.Logger) fx.Option {
	return fx.Provide(func() (out OutOption) {
		out.Option = func(options *Options) error {
			options.ErrorLog = logger
			return nil
		}
		return
	})
}

func ReadTimeout(t time.Duration) fx.Option {
	return fx.Provide(func() (out OutOption) {
		out.Option = func(options *Options) error {
			options.ReadTimeout = t
			return nil
		}
		return
	})
}

func ReadHeaderTimeout(t time.Duration) fx.Option {
	return fx.Provide(func() (out OutOption) {
		out.Option = func(options *Options) error {
			options.ReadHeaderTimeout = t
			return nil
		}
		return
	})
}

func WriteTimeout(t time.Duration) fx.Option {
	return fx.Provide(func() (out OutOption) {
		out.Option = func(options *Options) error {
			options.WriteTimeout = t
			return nil
		}
		return
	})
}
func IdleTimeout(t time.Duration) fx.Option {
	return fx.Provide(func() (out OutOption) {
		out.Option = func(options *Options) error {
			options.IdleTimeout = t
			return nil
		}
		return
	})
}

func MaxHeaderBytes(s int) fx.Option {
	return fx.Provide(func() (out OutOption) {
		out.Option = func(options *Options) error {
			options.MaxHeaderBytes = s
			return nil
		}
		return
	})
}
func TLSConfig(s *tls.Config) fx.Option {
	return fx.Provide(func() (out OutOption) {
		out.Option = func(options *Options) error {
			options.TLSConfig = s
			return nil
		}
		return
	})
}

func ShutdownTimeout(b time.Duration) fx.Option {
	return fx.Provide(func() (out OutOption) {
		out.Option = func(options *Options) error {
			options.ShutdownTimeout = b
			return nil
		}
		return
	})
}

func Handler(hosts []string, index int, handler HandlerFunc) fx.Option {
	return fx.Provide(func() (out OutOption) {
		out.Option = func(options *Options) error {
			options.Handlers = append(options.Handlers, HandlerOption{Hosts: hosts, Index: index, Handler: handler})
			return nil
		}
		return
	})
}

func DefaultOptions() *Options {
	return &Options{
		Addr:                   8080,
		ReadTimeout:            time.Second * 18000,
		ReadHeaderTimeout:      time.Second * 10,
		WriteTimeout:           time.Second * 18000,
		IdleTimeout:            time.Second * 1800,
		MaxHeaderBytes:         4096,
		ShutdownTimeout:        time.Hour,
		ShutdownRequestTimeout: time.Second * 15,
	}
}
