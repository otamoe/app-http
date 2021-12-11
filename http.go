package apphttp

import (
	"context"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/otamoe/app-http/certificate"

	"go.uber.org/fx"
)

func New(addr string) (options fx.Option) {
	options = fx.Options(
		fx.Provide(
			func(ctx context.Context) (options *Options) {
				options = DefaultOptions()
				return
			}),
		fx.Provide(func(options *Options, in InOptions, lc fx.Lifecycle) (server *http.Server, err error) {
			for _, v := range in.Options {
				if err = v(options); err != nil {
					return
				}
			}
			server = &http.Server{
				ErrorLog:          options.ErrorLog,
				ReadTimeout:       options.ReadTimeout,
				ReadHeaderTimeout: options.ReadHeaderTimeout,
				WriteTimeout:      options.WriteTimeout,
				IdleTimeout:       options.IdleTimeout,
				MaxHeaderBytes:    options.MaxHeaderBytes,
				TLSConfig:         options.TLSConfig,
				Addr:              options.Addr,
			}
			if server.Addr == "" {
				if server.TLSConfig == nil {
					server.Addr = ":8080"
				} else {
					server.Addr = ":8443"
				}
			}

			if server.TLSConfig == nil && (server.Addr == ":443" || server.Addr == ":8443") {
				var cert *certificate.Certificate
				if cert, err = certificate.CreateTLSCertificate("ecdsa", 384, "localhost", []string{"localhost"}, false, nil); err != nil {
					return
				}

				if server.TLSConfig, err = certificate.TLSConfig([]*certificate.Certificate{cert}); err != nil {
					return
				}
			}

			shutdownContext, shutdownCancel := context.WithCancel(context.Background())

			// 控制器 未找到
			notFoundHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			})

			// 控制器 排序
			sort.Slice(options.Handlers, func(i, j int) bool {
				return options.Handlers[i].Index < options.Handlers[j].Index
			})

			// 中间件 倒序 最后面添加的 丢到最里面去
			handlers := map[string][]HandlerFunc{}

			for _, oh := range options.Handlers {

				if len(oh.Hosts) == 0 {
					oh.Hosts = []string{"*"}
				}
				for _, host := range oh.Hosts {
					host = strings.ToLower(host)
					if host == "*" || host == "" {
						// 全局控制器
						for k, _ := range handlers {
							handlers[k] = append(handlers[k], oh.Handler)
						}

						// 没 * 写入到 *
						if _, ok := handlers[""]; !ok {
							handlers[""] = []HandlerFunc{}
							handlers[""] = append(handlers[""], oh.Handler)
						}
					} else {
						// 局部控制器
						if _, ok := handlers[host]; !ok {
							handlers[host] = []HandlerFunc{}

							// 添加 * host 的
							if _, ok := handlers[""]; ok {
								handlers[host] = append(handlers[host], handlers[""]...)
							}

						}
						handlers[host] = append(handlers[host], oh.Handler)
					}
				}
			}

			httpHandlers := map[string]http.Handler{}
			for host, value := range handlers {
				// 中间件 倒序 最后面添加的 丢到最里面去
				var httpHandler http.Handler
				httpHandler = notFoundHandler

				for i := len(value) - 1; i >= 0; i-- {
					httpHandler = value[i](httpHandler)
				}

				httpHandlers[host] = httpHandler
			}

			// 控制器
			server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

				// 添加取消上下文
				ctx, cancel := context.WithCancel(r.Context())
				defer cancel()

				go func() {
					select {
					case <-shutdownContext.Done():
						// 收到下线请求 取消当前链接
						cancel()
					case <-ctx.Done():
						// 请求已结束
					}
				}()

				r = r.WithContext(ctx)
				key := Host(r, "")
				if handler, ok := httpHandlers[key]; ok {
					handler.ServeHTTP(w, r)
				} else if handler, ok := httpHandlers[""]; ok {
					handler.ServeHTTP(w, r)
				} else {
					notFoundHandler.ServeHTTP(w, r)
				}
			})

			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) (err error) {
					t := time.NewTimer(3 * time.Second)
					defer t.Stop()
					errc := make(chan error, 1)
					go func() {
						var e error
						if server.TLSConfig == nil {
							e = server.ListenAndServe()
						} else {
							e = server.ListenAndServeTLS("", "")
						}
						if e == http.ErrServerClosed {
							e = nil
						}
						errc <- e
					}()
					select {
					case <-t.C:
					case err = <-errc:
					}
					return
				},
				OnStop: func(ctx context.Context) error {

					// 下线总共的超时
					if options.ShutdownTimeout != 0 {
						var cancel context.CancelFunc
						ctx, cancel = context.WithTimeout(ctx, options.ShutdownTimeout)
						defer cancel()
					}

					//  下线流
					done := make(chan struct{})

					// 下线
					go func() {
						err = server.Shutdown(ctx)
						close(done)
					}()

					// 下线请求超时
					if options.ShutdownRequestTimeout != 0 {
						timer := time.NewTimer(options.ShutdownRequestTimeout)
						defer timer.Stop()
						select {
						case <-timer.C:
							shutdownCancel()
						case <-done:
						}
					}

					//已下线
					<-done

					shutdownCancel()

					return err
				},
			})
			return
		}),
	)
	return
}

func Host(r *http.Request, defaultValue string) (host string) {
	if host = r.Header.Get("X-Forwarded-Host"); host != "" {

	} else if host = r.Host; host != "" {

	} else if host = r.Header.Get("X-Host"); host != "" {

	} else if host = r.URL.Host; host != "" {

	} else {
		host = defaultValue
	}

	if u, err := url.Parse("http://" + host); err == nil {
		return u.Hostname()
	}
	return host
}
