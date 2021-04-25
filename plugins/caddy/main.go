package caddy

import (
	"bytes"
	"context"
	"fmt"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/darkweak/souin/cache/coalescing"
	"github.com/darkweak/souin/configurationtypes"
	"github.com/darkweak/souin/plugins"
	"github.com/darkweak/souin/rfc"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
)

type key string

const getterContextCtxKey key = "getter_context"
const moduleName = "souin_cache"
const moduleID caddy.ModuleID = "http.handlers." + moduleName

func init() {
	caddy.RegisterModule(SouinCaddyPlugin{})
	httpcaddyfile.RegisterGlobalOption(moduleName, parseCaddyfileGlobalOption)
	httpcaddyfile.RegisterHandlerDirective(moduleName, parseCaddyfileHandlerDirective)
}

var (
	staticConfig Configuration
	appCounter = 0
	appConfigs []*Configuration
)

// SouinCaddyPlugin declaration.
type SouinCaddyPlugin struct {
	plugins.SouinBasePlugin
	Configuration *Configuration
	logger        *zap.Logger
	bufPool       sync.Pool
	Distributed   bool
	Headers       []string
	Olric         configurationtypes.CacheProvider
	TTL           string
	Rules         map[string]configurationtypes.URL
}

// CaddyModule returns the Caddy module information.
func (SouinCaddyPlugin) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  moduleID,
		New: func() caddy.Module { return new(SouinCaddyPlugin) },
	}
}

type getterContext struct {
	rw   http.ResponseWriter
	req  *http.Request
	next caddyhttp.Handler
}

// ServeHTTP implements caddyhttp.MiddlewareHandler.
func (s SouinCaddyPlugin) ServeHTTP(rw http.ResponseWriter, req *http.Request, next caddyhttp.Handler) error {
	getterCtx := getterContext{rw, req, next}
	ctx := context.WithValue(req.Context(), getterContextCtxKey, getterCtx)
	req = req.WithContext(ctx)
	buf := s.bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer s.bufPool.Put(buf)
	combo := ctx.Value(getterContextCtxKey).(getterContext)
	fmt.Println(s.Configuration.DefaultCache)
	plugins.DefaultSouinPluginCallback(rw, req, s.Retriever, s.RequestCoalescing, func(_ http.ResponseWriter, _ *http.Request) error {
		recorder := httptest.NewRecorder()
		e := combo.next.ServeHTTP(recorder, combo.req)
		if e != nil {
			return e
		}

		response := recorder.Result()
		req.Response = response
		response, e = s.Retriever.GetTransport().(*rfc.VaryTransport).UpdateCacheEventually(req)
		if e != nil {
			return e
		}

		_, e = io.Copy(rw, response.Body)

		return e
	})

	return nil
}

// Validate to validate configuration.
func (s *SouinCaddyPlugin) Validate() error {
	return nil
}

// Provision to do the provisioning part.
func (s *SouinCaddyPlugin) Provision(ctx caddy.Context) error {
	s.logger = ctx.Logger(s)

	s.bufPool = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
	if s.Configuration == nil && appConfigs[appCounter] != nil {
		s.Configuration = appConfigs[appCounter]
	}
	s.Retriever = plugins.DefaultSouinPluginInitializerFromConfiguration(s.Configuration)
	s.RequestCoalescing = coalescing.Initialize()
	appCounter += 1
	return nil
}

func parseCaddyfileGlobalOption(h *caddyfile.Dispenser) (interface{}, error) {
	var souin SouinCaddyPlugin
	cfg := &Configuration{
		DefaultCache: &DefaultCache{
			Distributed: false,
			Headers: []string{},
		},
		URLs: make(map[string]configurationtypes.URL),
	}

	for h.Next() {
		for nesting := h.Nesting(); h.NextBlock(nesting); {
			rootOption := h.Val()
			switch rootOption {
			case "distributed":
				args := h.RemainingArgs()
				distributed, _ := strconv.ParseBool(args[0])
				cfg.DefaultCache.Distributed = distributed
			case "headers":
				args := h.RemainingArgs()
				cfg.DefaultCache.Headers = append(cfg.DefaultCache.Headers, args...)
			case "log_level":
				args := h.RemainingArgs()
				cfg.LogLevel = args[0]
			case "olric":
				provider := configurationtypes.CacheProvider{}
				for nesting := h.Nesting(); h.NextBlock(nesting); {
					directive := h.Val()
					switch directive {
					case "url":
						urlArgs := h.RemainingArgs()
						provider.URL = urlArgs[0]
					}
				}
				cfg.DefaultCache.Olric = provider
			case "ttl":
				args := h.RemainingArgs()
				cfg.DefaultCache.TTL = args[0]
			default:
				return nil, h.Errf("unsupported root directive: %s", rootOption)
			}
		}
	}

	souin.Configuration = cfg
	staticConfig = *cfg
	return nil, nil
}

func parseCaddyfileHandlerDirective(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	dc := DefaultCache{
		Distributed: staticConfig.DefaultCache.Distributed,
		Headers:     staticConfig.DefaultCache.Headers,
		Olric:       staticConfig.DefaultCache.Olric,
		Regex:       staticConfig.DefaultCache.Regex,
		TTL:         staticConfig.DefaultCache.TTL,
	}
	sc := Configuration{
		DefaultCache: &dc,
		URLs:         staticConfig.URLs,
		LogLevel:     staticConfig.LogLevel,
		logger:       staticConfig.logger,
	}

	for h.Next() {
		directive := h.Val()
		switch directive {
		case "headers":
			sc.DefaultCache.Headers = h.RemainingArgs()
		case "ttl":
			sc.DefaultCache.TTL = h.RemainingArgs()[0]
		}
	}

	if appConfigs == nil || len(appConfigs) == 0 {
		appConfigs = []*Configuration{&sc}
	} else {
		appConfigs = append(appConfigs, &sc)
	}

	return &SouinCaddyPlugin{
		Configuration: &sc,
	}, nil
}

// Interface guards
var (
	_ caddy.Provisioner           = (*SouinCaddyPlugin)(nil)
	_ caddyhttp.MiddlewareHandler = (*SouinCaddyPlugin)(nil)
	_ caddy.Validator             = (*SouinCaddyPlugin)(nil)
)
