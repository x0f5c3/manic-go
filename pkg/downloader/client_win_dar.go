//go:build windows || (darwin && arm64)

package downloader

import (
	"strings"
	"time"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpproxy"
)

//	var DefaultClient = fasthttp.Client{
//		Name:                     "manic-client",
//		NoDefaultUserAgentHeader: true,
//		Dial:                     defaultDial,
//		DialDualStack:            true,
//		ConfigureClient:          nil,
//	}
var defaultDial fasthttp.DialFunc = fasthttp.Dial

func ProxyDialer(proxy string, timeout time.Duration) fasthttp.DialFunc {
	if strings.Contains(proxy, "socks5:") {
		return fasthttpproxy.FasthttpSocksDialer(proxy)
	}
	return fasthttpproxy.FasthttpHTTPDialerTimeout(proxy, timeout)
}
