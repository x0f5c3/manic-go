package downloader

import (
	"encoding/json"
	"net"

	"github.com/valyala/fasthttp"
)

type Client struct {
	*fasthttp.Client
}

var DefaultClient = NewClient()

func NewClient() *Client {
	cl := &fasthttp.Client{
		Name:                          "manic-client",
		NoDefaultUserAgentHeader:      true,
		DisableHeaderNamesNormalizing: false,
		DisablePathNormalizing:        false,
		Dial:                          defaultDial,
		DialDualStack:                 true,
	}
	return &Client{Client: cl}
}

type Addr struct {
	network string
	addr    string
}

func (a Addr) Network() string {
	return a.network
}

func (a Addr) String() string {
	return a.addr
}

func fromAddr(addr net.Addr) Addr {
	return Addr{
		network: addr.Network(),
		addr:    addr.String(),
	}
}

// type Response struct {
// 	Headers map[string]string
// 	Body    []byte
// 	laddr   net.Addr
// 	raddr   net.Addr
// }
//
// func fromResponse(response *fasthttp.Response) (*Response, error) {
// 	heads := make(map[string]string)
// 	response.Header.VisitAll(func(key, value []byte) {
// 		heads[string(key)] = string(value)
// 	})
// 	var body []byte
// 	bw := response.BodyWriter()
// 	_, err := bw.Write(body)
// 	if err != nil {
// 		return nil, err
// 	}
// 	laddr := fromAddr(response.LocalAddr())
// 	raddr := fromAddr(response.RemoteAddr())
// 	return &Response{
// 		Headers: heads,
// 		Body:    body,
// 		laddr:   laddr,
// 		raddr:   raddr,
// 	}, nil
// }

type Response struct {
	*fasthttp.Response
}

func (r *Response) Json(res any) error {
	return json.Unmarshal(r.Body(), res)
}

func AcquireResponse() *Response {
	return &Response{Response: fasthttp.AcquireResponse()}
}
func ReleaseResponse(response *Response) {
	fasthttp.ReleaseResponse(response.Response)
}

func (c *Client) Head(url string) (*Response, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	req.SetRequestURI(url)
	req.Header.SetMethod("HEAD")
	resp := AcquireResponse()
	err := c.DoRedirects(req, resp.Response, 30)
	if err != nil {
		return nil, err
	}
	// res, err := fromResponse(resp)
	// if err != nil {
	// 	return nil, err
	// }
	return resp, nil
}

func (c *Client) GetRange(url, val string) (*Response, error) {
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(url)
	req.Header.SetMethod("GET")
	req.Header.Set("RANGE", val)
	resp := AcquireResponse()
	err := c.DoRedirects(req, resp.Response, 30)
	if err != nil {
		return nil, err
	}
	fasthttp.ReleaseRequest(req)
	return resp, nil
}

func (c *Client) Get(url string) (*Response, error) {
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(url)
	req.Header.SetMethod("GET")
	resp := AcquireResponse()
	err := c.DoRedirects(req, resp.Response, 30)
	if err != nil {
		return nil, err
	}
	fasthttp.ReleaseRequest(req)

	return resp, nil
}

func Get(url string) (*Response, error) {
	return DefaultClient.Get(url)
}
