package mihomo

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/evilCYH/NodeHub/internal/database/op"
	"github.com/evilCYH/NodeHub/internal/models/setting"
	"github.com/evilCYH/NodeHub/internal/utils/log"
	"github.com/metacubex/mihomo/adapter"
	"github.com/metacubex/mihomo/constant"
)

type HC struct {
	*http.Client
	proxy constant.Proxy
}

var clientPool = sync.Pool{
	New: func() interface{} {
		return &http.Client{
			Timeout: 300 * time.Second,
		}
	},
}

var transportPool = sync.Pool{
	New: func() interface{} {
		return &http.Transport{
			DisableKeepAlives:     false,
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   10,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   30 * time.Second,
			ExpectContinueTimeout: 10 * time.Second,
			ResponseHeaderTimeout: 30 * time.Second,
		}
	},
}

func parsePort(portStr string) (uint16, error) {
	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return 0, err
	}
	return uint16(port), nil
}

func Default(useProxy bool) *HC {
	if !useProxy || !op.GetSettingBool(setting.PROXY_ENABLE) {
		return direct()
	}
	proxyUrl := op.GetSettingStr(setting.PROXY_URL)
	if proxyUrl == "" {
		log.Warnf("proxy url is empty")
		return direct()
	}

	parsed, err := url.Parse(proxyUrl)
	if err != nil {
		log.Warnf("parse proxy url failed: %v", err)
		return direct()
	}

	host, portStr, err := net.SplitHostPort(parsed.Host)
	if err != nil {
		log.Warnf("split host port failed: %v", err)
		return direct()
	}

	portInt, err := parsePort(portStr)
	if err != nil {
		log.Warnf("parse port failed: %v", err)
		return direct()
	}

	proxyConfig := map[string]any{
		"name":     "proxy",
		"server":   host,
		"port":     portInt,
		"username": parsed.User.Username(),
	}
	if password, ok := parsed.User.Password(); ok {
		proxyConfig["password"] = password
	}
	switch parsed.Scheme {
	case "socks5":
		proxyConfig["type"] = "socks5"
	case "http":
		proxyConfig["type"] = "http"
	case "https":
		proxyConfig["type"] = "http"
		proxyConfig["tls"] = true
	default:
		log.Warnf("unsupported proxy scheme: %s", parsed.Scheme)
		return direct()
	}
	return Proxy(proxyConfig)
}

func direct() *HC {
	var directProxy = map[string]any{
		"name": "direct",
		"type": "direct",
	}
	return Proxy(directProxy)
}

func Proxy(raw map[string]any) *HC {
	if raw == nil {
		log.Warnf("proxy config is nil")
		return nil
	}
	proxy, err := adapter.ParseProxy(raw)
	if err != nil {
		if proxy != nil {
			proxy.Close()
		}
		log.Debugf("parse proxy failed: %v raw: %v", err, raw)
		return nil
	}

	client := clientPool.Get().(*http.Client)
	transport := transportPool.Get().(*http.Transport)

	transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		host, portStr, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, err
		}

		u16Port, err := parsePort(portStr)
		if err != nil {
			log.Warnf("parse port failed, using port 0: %v", err)
			u16Port = 0
		}

		return proxy.DialContext(ctx, &constant.Metadata{
			Host:    host,
			DstPort: u16Port,
		})
	}

	client.Transport = transport
	return &HC{Client: client, proxy: proxy}
}

func (h *HC) Release() {
	if h.Client == nil {
		return
	}
	if h.proxy != nil {
		h.proxy.Close()
		h.proxy = nil
	}
	if transport, ok := h.Transport.(*http.Transport); ok {
		transport.DialContext = nil
		transport.TLSClientConfig = nil
		transport.Proxy = nil
		transport.CloseIdleConnections()
		transportPool.Put(transport)
	}
	h.Transport = nil
	h.Timeout = 300 * time.Second
	h.CheckRedirect = nil
	h.Jar = nil
	clientPool.Put(h.Client)
}
