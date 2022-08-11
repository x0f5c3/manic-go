package downloader

import "github.com/valyala/fasthttp"

// var DefaultClient = fasthttp.Client{
// 	Name:                     "manic-client",
// 	NoDefaultUserAgentHeader: true,
// 	Dial:                     defaultDial,
// 	DialDualStack:            true,
// 	ConfigureClient:          nil,
// }
var defaultDial fasthttp.DialFunc = fasthttp.Dial
